package postgres

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"api-go/internal/inventory/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inventoryBatchModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProductID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	BatchNumber    string     `gorm:"type:varchar(100);not null"`
	ExpirationDate *time.Time `gorm:"type:date"`
	Quantity       int        `gorm:"not null"`
	Cost           *float64   `gorm:"type:decimal(10,2)"`
	SupplierID     *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
}

func (inventoryBatchModel) TableName() string {
	return "inventory_batches"
}

type inventoryLogModel struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductName   string    `gorm:"type:varchar(255);not null"`
	VariantID     *uuid.UUID
	VariantName   *string `gorm:"type:varchar(255)"`
	BatchID       *uuid.UUID
	BatchNumber   *string   `gorm:"type:varchar(100)"`
	SKU           string    `gorm:"type:varchar(100);not null"`
	Type          string    `gorm:"type:varchar(50);not null"`
	Quantity      int       `gorm:"not null"`
	PreviousStock int       `gorm:"not null"`
	NewStock      int       `gorm:"not null"`
	Reason        string    `gorm:"type:text;not null"`
	ReasonType    *string   `gorm:"type:varchar(100)"`
	UserID        string    `gorm:"type:varchar(100);not null"`
	UserName      string    `gorm:"type:varchar(255);not null"`
	Notes         *string   `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null;index"`
	Undoable      bool      `gorm:"not null;default:true"`
}

func (inventoryLogModel) TableName() string {
	return "inventory_logs"
}

type stockAdjustmentModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProductID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	ProductName string     `gorm:"type:varchar(255);not null"`
	BatchID     *uuid.UUID `gorm:"type:uuid"`
	BatchNumber *string    `gorm:"type:varchar(100)"`
	SKU         string     `gorm:"type:varchar(100);not null"`
	OldStock    int        `gorm:"not null"`
	NewStock    int        `gorm:"not null"`
	Reason      string     `gorm:"type:text;not null"`
	UserID      string     `gorm:"type:varchar(100);not null"`
	UserName    string     `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time  `gorm:"not null"`
}

func (stockAdjustmentModel) TableName() string {
	return "stock_adjustments"
}

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return &inventoryRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&inventoryBatchModel{},
		&inventoryLogModel{},
		&stockAdjustmentModel{},
	)
}

func (r *inventoryRepository) GetBatchesByProductID(ctx context.Context, productID uuid.UUID) ([]*domain.InventoryBatch, error) {
	var models []inventoryBatchModel
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("created_at ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	batches := make([]*domain.InventoryBatch, len(models))
	for i, m := range models {
		batches[i] = toBatchDomain(&m)
	}
	return batches, nil
}

func (r *inventoryRepository) GetAllBatches(ctx context.Context) ([]*domain.InventoryBatch, error) {
	var models []inventoryBatchModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	batches := make([]*domain.InventoryBatch, len(models))
	for i, m := range models {
		batches[i] = toBatchDomain(&m)
	}
	return batches, nil
}

func (r *inventoryRepository) GetBatchByID(ctx context.Context, id uuid.UUID) (*domain.InventoryBatch, error) {
	var model inventoryBatchModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBatchNotFound
		}
		return nil, err
	}
	return toBatchDomain(&model), nil
}

func (r *inventoryRepository) CreateBatch(ctx context.Context, batch *domain.InventoryBatch) error {
	model := toBatchModel(batch)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *inventoryRepository) UpdateBatchQuantity(ctx context.Context, id uuid.UUID, quantity int) error {
	return r.db.WithContext(ctx).Model(&inventoryBatchModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"quantity":   quantity,
		"updated_at": time.Now(),
	}).Error
}

func (r *inventoryRepository) GetAvailableBatchesForProduct(ctx context.Context, productID uuid.UUID) ([]*domain.InventoryBatch, error) {
	var models []inventoryBatchModel
	err := r.db.WithContext(ctx).Where("product_id = ? AND quantity > 0", productID).Order("created_at ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	batches := make([]*domain.InventoryBatch, len(models))
	for i, m := range models {
		batches[i] = toBatchDomain(&m)
	}
	return batches, nil
}

func (r *inventoryRepository) CreateLog(ctx context.Context, log *domain.InventoryLog) error {
	model := toLogModel(log)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *inventoryRepository) GetLogByID(ctx context.Context, id uuid.UUID) (*domain.InventoryLog, error) {
	var model inventoryLogModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogNotFound
		}
		return nil, err
	}
	return toLogDomain(&model), nil
}

func (r *inventoryRepository) UpdateLogUndoable(ctx context.Context, id uuid.UUID, undoable bool) error {
	return r.db.WithContext(ctx).Model(&inventoryLogModel{}).Where("id = ?", id).Update("undoable", undoable).Error
}

func (r *inventoryRepository) GetLogs(ctx context.Context, filters *domain.InventoryFilters) (*domain.PaginatedLogsResult, error) {
	var models []inventoryLogModel
	var total int64

	query := r.db.WithContext(ctx).Model(&inventoryLogModel{})

	if filters != nil {
		if filters.Type != nil && *filters.Type != "" {
			query = query.Where("type = ?", *filters.Type)
		}
		if filters.ProductID != nil && *filters.ProductID != uuid.Nil {
			query = query.Where("product_id = ?", *filters.ProductID)
		}
		if filters.BatchID != nil && *filters.BatchID != uuid.Nil {
			query = query.Where("batch_id = ?", *filters.BatchID)
		}
		if filters.UserID != nil && *filters.UserID != "" {
			query = query.Where("user_id = ?", *filters.UserID)
		}
		if filters.ReasonType != nil && *filters.ReasonType != "" {
			query = query.Where("reason_type = ?", *filters.ReasonType)
		}
		if filters.StartDate != nil {
			query = query.Where("created_at >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			query = query.Where("created_at <= ?", *filters.EndDate)
		}
		if filters.Search != "" {
			search := "%" + strings.ToLower(filters.Search) + "%"
			query = query.Where("LOWER(product_name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(user_name) LIKE ? OR LOWER(batch_number) LIKE ?", search, search, search, search)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	page := 1
	limit := 20
	if filters != nil {
		if filters.Page > 0 {
			page = filters.Page
		}
		if filters.Limit > 0 {
			limit = filters.Limit
		}
	}

	sortBy := "created_at"
	sortOrder := "DESC"
	if filters != nil {
		if filters.SortBy != "" {
			sortBy = filters.SortBy
		}
		if strings.ToUpper(filters.SortOrder) == "ASC" {
			sortOrder = "ASC"
		}
	}

	offset := (page - 1) * limit
	err := query.Order(sortBy + " " + sortOrder).Offset(offset).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}

	logs := make([]*domain.InventoryLog, len(models))
	for i, m := range models {
		logs[i] = toLogDomain(&m)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &domain.PaginatedLogsResult{
		Data:       logs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (r *inventoryRepository) CreateStockAdjustment(ctx context.Context, adj *domain.StockAdjustment) error {
	model := toStockAdjustmentModel(adj)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *inventoryRepository) GetCriticalStock(ctx context.Context, minStock int) ([]*domain.CriticalStockItem, error) {
	type resultStruct struct {
		ID          uuid.UUID
		Name        string
		SKU         string
		Quantity    int
		MinQuantity int
		Category    string
	}

	var results []resultStruct
	err := r.db.WithContext(ctx).Table("products").
		Select("id, name, sku, quantity, min_quantity, category").
		Where("quantity <= min_quantity OR quantity <= ?", minStock).
		Order("quantity ASC").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	items := make([]*domain.CriticalStockItem, len(results))
	for i, res := range results {
		catName := res.Category
		items[i] = &domain.CriticalStockItem{
			ProductID:    res.ID,
			ProductName:  res.Name,
			SKU:          res.SKU,
			CurrentStock: res.Quantity,
			MinStock:     res.MinQuantity,
			CategoryName: &catName,
		}
	}

	return items, nil
}

func (r *inventoryRepository) GetMetrics(ctx context.Context) (*domain.InventoryMetrics, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var totalProducts int64
	r.db.WithContext(ctx).Table("products").Count(&totalProducts)

	var outOfStock int64
	r.db.WithContext(ctx).Table("products").Where("quantity <= 0").Count(&outOfStock)

	var criticalItems int64
	r.db.WithContext(ctx).Table("products").Where("quantity > 0 AND quantity <= min_quantity").Count(&criticalItems)

	var totalValue float64
	r.db.WithContext(ctx).Table("products").Select("COALESCE(SUM(cost * quantity), 0)").Scan(&totalValue)

	var inboundThisMonth int64
	r.db.WithContext(ctx).Model(&inventoryLogModel{}).
		Where("type = ? AND created_at >= ?", domain.LogTypeInbound, startOfMonth).
		Select("COALESCE(SUM(quantity), 0)").Scan(&inboundThisMonth)

	var outboundThisMonth int64
	r.db.WithContext(ctx).Model(&inventoryLogModel{}).
		Where("type = ? AND created_at >= ?", domain.LogTypeOutbound, startOfMonth).
		Select("COALESCE(SUM(quantity), 0)").Scan(&outboundThisMonth)

	var adjustmentsThisMonth int64
	r.db.WithContext(ctx).Model(&inventoryLogModel{}).
		Where("type = ? AND created_at >= ?", domain.LogTypeAdjustment, startOfMonth).
		Count(&adjustmentsThisMonth)

	return &domain.InventoryMetrics{
		TotalValue:           totalValue,
		TotalProducts:        totalProducts,
		CriticalItems:        criticalItems,
		OutOfStock:           outOfStock,
		InboundThisMonth:     inboundThisMonth,
		OutboundThisMonth:    outboundThisMonth,
		AdjustmentsThisMonth: adjustmentsThisMonth,
	}, nil
}

// Helpers

func toBatchDomain(m *inventoryBatchModel) *domain.InventoryBatch {
	return &domain.InventoryBatch{
		ID:             m.ID,
		ProductID:      m.ProductID,
		BatchNumber:    m.BatchNumber,
		ExpirationDate: m.ExpirationDate,
		Quantity:       m.Quantity,
		Cost:           m.Cost,
		SupplierID:     m.SupplierID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func toBatchModel(b *domain.InventoryBatch) *inventoryBatchModel {
	return &inventoryBatchModel{
		ID:             b.ID,
		ProductID:      b.ProductID,
		BatchNumber:    b.BatchNumber,
		ExpirationDate: b.ExpirationDate,
		Quantity:       b.Quantity,
		Cost:           b.Cost,
		SupplierID:     b.SupplierID,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
	}
}

func toLogDomain(m *inventoryLogModel) *domain.InventoryLog {
	return &domain.InventoryLog{
		ID:            m.ID,
		ProductID:     m.ProductID,
		ProductName:   m.ProductName,
		VariantID:     m.VariantID,
		VariantName:   m.VariantName,
		BatchID:       m.BatchID,
		BatchNumber:   m.BatchNumber,
		SKU:           m.SKU,
		Type:          domain.InventoryLogType(m.Type),
		Quantity:      m.Quantity,
		PreviousStock: m.PreviousStock,
		NewStock:      m.NewStock,
		Reason:        m.Reason,
		ReasonType:    m.ReasonType,
		UserID:        m.UserID,
		UserName:      m.UserName,
		Notes:         m.Notes,
		CreatedAt:     m.CreatedAt,
		Undoable:      m.Undoable,
	}
}

func toLogModel(l *domain.InventoryLog) *inventoryLogModel {
	return &inventoryLogModel{
		ID:            l.ID,
		ProductID:     l.ProductID,
		ProductName:   l.ProductName,
		VariantID:     l.VariantID,
		VariantName:   l.VariantName,
		BatchID:       l.BatchID,
		BatchNumber:   l.BatchNumber,
		SKU:           l.SKU,
		Type:          string(l.Type),
		Quantity:      l.Quantity,
		PreviousStock: l.PreviousStock,
		NewStock:      l.NewStock,
		Reason:        l.Reason,
		ReasonType:    l.ReasonType,
		UserID:        l.UserID,
		UserName:      l.UserName,
		Notes:         l.Notes,
		CreatedAt:     l.CreatedAt,
		Undoable:      l.Undoable,
	}
}

func toStockAdjustmentModel(s *domain.StockAdjustment) *stockAdjustmentModel {
	return &stockAdjustmentModel{
		ID:          s.ID,
		ProductID:   s.ProductID,
		ProductName: s.ProductName,
		BatchID:     s.BatchID,
		BatchNumber: s.BatchNumber,
		SKU:         s.SKU,
		OldStock:    s.OldStock,
		NewStock:    s.NewStock,
		Reason:      s.Reason,
		UserID:      s.UserID,
		UserName:    s.UserName,
		CreatedAt:   s.CreatedAt,
	}
}
