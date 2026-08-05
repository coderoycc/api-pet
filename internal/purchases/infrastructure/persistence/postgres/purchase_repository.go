package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"api-go/internal/purchases/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type purchaseModel struct {
	ID          uuid.UUID           `gorm:"type:uuid;primaryKey"`
	Number      string              `gorm:"type:varchar(50);not null;uniqueIndex"`
	SupplierID  uuid.UUID           `gorm:"type:uuid;not null"`
	WarehouseID uuid.UUID           `gorm:"type:uuid;not null"`
	Status      string              `gorm:"type:varchar(50);not null"`
	TotalAmount float64             `gorm:"type:decimal(10,2);not null"`
	Notes       *string             `gorm:"type:text"`
	CreatedAt   time.Time           `gorm:"not null"`
	ReceivedAt  *time.Time          
	UpdatedAt   time.Time           `gorm:"not null"`
	Items       []purchaseItemModel `gorm:"foreignKey:PurchaseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (purchaseModel) TableName() string {
	return "purchases"
}

type purchaseItemModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	PurchaseID uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null"`
	Quantity   int       `gorm:"not null"`
	UnitCost   float64   `gorm:"type:decimal(10,2);not null"`
	Subtotal   float64   `gorm:"type:decimal(10,2);not null"`
}

func (purchaseItemModel) TableName() string {
	return "purchase_items"
}

type purchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) domain.PurchaseRepository {
	return &purchaseRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	// Create the sequence for order numbers if it doesn't exist
	db.Exec("CREATE SEQUENCE IF NOT EXISTS purchase_number_seq START 1")

	return db.AutoMigrate(&purchaseModel{}, &purchaseItemModel{})
}

func (r *purchaseRepository) Create(ctx context.Context, purchase *domain.Purchase) error {
	var nextval int
	if err := r.db.WithContext(ctx).Raw("SELECT nextval('purchase_number_seq')").Scan(&nextval).Error; err != nil {
		return err
	}

	// Format: OC-YYYY-NNN
	purchase.Number = fmt.Sprintf("OC-%d-%03d", time.Now().Year(), nextval)

	model := toPurchaseModel(purchase)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	return nil
}

func (r *purchaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Purchase, error) {
	var model purchaseModel

	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&model, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPurchaseNotFound
		}
		return nil, err
	}

	return r.mapToDomainWithJoins(ctx, &model)
}

func (r *purchaseRepository) GetAll(ctx context.Context, page, limit int, filters *domain.PurchaseFilters, sortBy, sortOrder string) (*domain.PaginatedPurchaseResult, error) {
	var models []purchaseModel
	var total int64

	query := r.db.WithContext(ctx).Model(&purchaseModel{})

	if filters != nil {
		if filters.Search != "" {
			search := "%" + strings.ToLower(filters.Search) + "%"
			query = query.Joins("LEFT JOIN suppliers ON suppliers.id = purchases.supplier_id").
				Joins("LEFT JOIN warehouses ON warehouses.id = purchases.warehouse_id").
				Where("LOWER(purchases.number) LIKE ? OR LOWER(suppliers.name) LIKE ? OR LOWER(warehouses.name) LIKE ?", search, search, search)
		}
		if filters.Status != "" {
			query = query.Where("purchases.status = ?", filters.Status)
		}
		if filters.SupplierID != nil {
			query = query.Where("purchases.supplier_id = ?", *filters.SupplierID)
		}
		if filters.WarehouseID != nil {
			query = query.Where("purchases.warehouse_id = ?", *filters.WarehouseID)
		}
		if filters.StartDate != "" {
			query = query.Where("purchases.created_at >= ?", filters.StartDate)
		}
		if filters.EndDate != "" {
			query = query.Where("purchases.created_at <= ?", filters.EndDate)
		}
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	if sortBy != "" {
		order := "ASC"
		if strings.ToLower(sortOrder) == "desc" {
			order = "DESC"
		}
		// prepend table name to prevent ambiguity
		query = query.Order("purchases." + sortBy + " " + order)
	} else {
		query = query.Order("purchases.created_at DESC")
	}

	offset := (page - 1) * limit
	err = query.Preload("Items").Offset(offset).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}

	purchases := make([]*domain.Purchase, len(models))
	for i, m := range models {
		mapped, err := r.mapToDomainWithJoins(ctx, &m)
		if err != nil {
			return nil, err
		}
		purchases[i] = mapped
	}

	return &domain.PaginatedPurchaseResult{
		Data:       purchases,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

func (r *purchaseRepository) UpdateStatus(ctx context.Context, purchase *domain.Purchase) error {
	updates := map[string]interface{}{
		"status":     string(purchase.Status),
		"updated_at": purchase.UpdatedAt,
	}
	
	if purchase.ReceivedAt != nil {
		updates["received_at"] = *purchase.ReceivedAt
	}

	result := r.db.WithContext(ctx).Model(&purchaseModel{}).Where("id = ?", purchase.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPurchaseNotFound
	}
	return nil
}

func (r *purchaseRepository) GetMetrics(ctx context.Context) (*domain.PurchaseMetrics, error) {
	var metrics domain.PurchaseMetrics

	r.db.WithContext(ctx).Model(&purchaseModel{}).Count(&metrics.TotalOrders)
	
	r.db.WithContext(ctx).Model(&purchaseModel{}).
		Where("status IN ?", []string{string(domain.PurchaseStatusPending), string(domain.PurchaseStatusConfirmed)}).
		Count(&metrics.PendingOrders)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	r.db.WithContext(ctx).Model(&purchaseModel{}).
		Where("status = ? AND received_at >= ?", string(domain.PurchaseStatusReceived), startOfMonth).
		Count(&metrics.ReceivedThisMonth)

	var totalAmount *float64
	r.db.WithContext(ctx).Model(&purchaseModel{}).
		Where("status = ? AND received_at >= ?", string(domain.PurchaseStatusReceived), startOfMonth).
		Select("SUM(total_amount)").Scan(&totalAmount)

	if totalAmount != nil {
		metrics.TotalAmountThisMonth = *totalAmount
	}

	return &metrics, nil
}

// Helpers

// mapToDomainWithJoins takes a model and performs necessary fast lookups to populate the read-only names
func (r *purchaseRepository) mapToDomainWithJoins(ctx context.Context, m *purchaseModel) (*domain.Purchase, error) {
	purchase := toPurchaseDomain(m)

	// Fetch Supplier Name
	var supplierName string
	r.db.WithContext(ctx).Table("suppliers").Select("name").Where("id = ?", purchase.SupplierID).Scan(&supplierName)
	purchase.SupplierName = supplierName

	// Fetch Warehouse Name
	var warehouseName string
	r.db.WithContext(ctx).Table("warehouses").Select("name").Where("id = ?", purchase.WarehouseID).Scan(&warehouseName)
	purchase.WarehouseName = warehouseName

	// Fetch Product Details for Items
	for i, item := range purchase.Items {
		type productDetails struct {
			Name string
			SKU  string
			Unit string
		}
		var pd productDetails
		r.db.WithContext(ctx).Table("products").Select("name", "sku", "unit").Where("id = ?", item.ProductID).Scan(&pd)
		purchase.Items[i].ProductName = pd.Name
		purchase.Items[i].SKU = pd.SKU
		purchase.Items[i].Unit = pd.Unit
	}

	return purchase, nil
}

func toPurchaseDomain(m *purchaseModel) *domain.Purchase {
	p := &domain.Purchase{
		ID:          m.ID,
		Number:      m.Number,
		SupplierID:  m.SupplierID,
		WarehouseID: m.WarehouseID,
		Status:      domain.PurchaseStatus(m.Status),
		TotalAmount: m.TotalAmount,
		Notes:       m.Notes,
		CreatedAt:   m.CreatedAt,
		ReceivedAt:  m.ReceivedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	if len(m.Items) > 0 {
		p.Items = make([]domain.PurchaseItem, len(m.Items))
		for i, item := range m.Items {
			p.Items[i] = domain.PurchaseItem{
				ID:         item.ID,
				PurchaseID: item.PurchaseID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				UnitCost:   item.UnitCost,
				Subtotal:   item.Subtotal,
			}
		}
	}

	return p
}

func toPurchaseModel(p *domain.Purchase) *purchaseModel {
	m := &purchaseModel{
		ID:          p.ID,
		Number:      p.Number,
		SupplierID:  p.SupplierID,
		WarehouseID: p.WarehouseID,
		Status:      string(p.Status),
		TotalAmount: p.TotalAmount,
		Notes:       p.Notes,
		CreatedAt:   p.CreatedAt,
		ReceivedAt:  p.ReceivedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if len(p.Items) > 0 {
		m.Items = make([]purchaseItemModel, len(p.Items))
		for i, item := range p.Items {
			m.Items[i] = purchaseItemModel{
				ID:         item.ID,
				PurchaseID: item.PurchaseID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				UnitCost:   item.UnitCost,
				Subtotal:   item.Subtotal,
			}
		}
	}

	return m
}
