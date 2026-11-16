package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"api-go/internal/sales/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type saleModel struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	InvoiceNumber string          `gorm:"type:varchar(50);not null;uniqueIndex"`
	CustomerID    *uuid.UUID      `gorm:"type:uuid"`
	CustomerName  string          `gorm:"type:varchar(255);not null"`
	UserID        uuid.UUID       `gorm:"type:uuid;not null"`
	UserName      string          `gorm:"type:varchar(255);not null"`
	TotalAmount   float64         `gorm:"type:decimal(10,2);not null"`
	PaymentMethod string          `gorm:"type:varchar(50);not null"`
	AmountPaid    float64         `gorm:"type:decimal(10,2);not null"`
	ChangeAmount  float64         `gorm:"type:decimal(10,2);not null"`
	Notes         string          `gorm:"type:text"`
	Items         []saleItemModel `gorm:"foreignKey:SaleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt     time.Time       `gorm:"not null"`
}

func (saleModel) TableName() string {
	return "sales"
}

type saleItemModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	SaleID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	ProductID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	ProductName      string     `gorm:"type:varchar(255);not null"`
	ProductSKU       string     `gorm:"type:varchar(100);not null"`
	ProductUnit      string     `gorm:"type:varchar(50);not null"`
	PresentationID   *uuid.UUID `gorm:"type:uuid"`
	PresentationName string     `gorm:"type:varchar(100)"`
	Quantity         float64    `gorm:"type:decimal(10,3);not null"`
	ConversionFactor float64    `gorm:"type:decimal(10,3);not null;default:1"`
	BaseQuantity     float64    `gorm:"type:decimal(12,3);not null"`
	UnitPrice        float64    `gorm:"type:decimal(10,2);not null"`
	Subtotal         float64    `gorm:"type:decimal(10,2);not null"`
}

func (saleItemModel) TableName() string {
	return "sale_items"
}

type saleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) domain.SaleRepository {
	return &saleRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	db.Exec("CREATE SEQUENCE IF NOT EXISTS sale_invoice_seq START 1")
	return db.AutoMigrate(&saleModel{}, &saleItemModel{})
}

func (r *saleRepository) Create(ctx context.Context, sale *domain.Sale) error {
	var nextval int
	if err := r.db.WithContext(ctx).Raw("SELECT nextval('sale_invoice_seq')").Scan(&nextval).Error; err != nil {
		return err
	}

	sale.InvoiceNumber = fmt.Sprintf("FAC-%d-%04d", time.Now().Year(), nextval)

	model := toSaleModel(sale)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *saleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Sale, error) {
	var model saleModel
	err := r.db.WithContext(ctx).Preload("Items").First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSaleNotFound
		}
		return nil, err
	}
	return toSaleDomain(&model), nil
}

func (r *saleRepository) GetAll(ctx context.Context, filters *domain.SaleFilters) (*domain.PaginatedSaleResult, error) {
	var models []saleModel
	var total int64

	query := r.db.WithContext(ctx).Model(&saleModel{})

	if filters != nil {
		if filters.Search != "" {
			search := "%" + strings.ToLower(filters.Search) + "%"
			query = query.Where("LOWER(invoice_number) LIKE ? OR LOWER(customer_name) LIKE ?", search, search)
		}
		if filters.PaymentMethod != "" {
			query = query.Where("payment_method = ?", filters.PaymentMethod)
		}
		if filters.CustomerID != nil {
			query = query.Where("customer_id = ?", *filters.CustomerID)
		}
		if filters.UserID != nil {
			query = query.Where("user_id = ?", *filters.UserID)
		}
		if filters.StartDate != nil {
			query = query.Where("created_at >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			query = query.Where("created_at <= ?", *filters.EndDate)
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

	offset := (page - 1) * limit
	err := query.Preload("Items").Order("created_at DESC").Offset(offset).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}

	sales := make([]*domain.Sale, len(models))
	for i, m := range models {
		sales[i] = toSaleDomain(&m)
	}

	return &domain.PaginatedSaleResult{
		Data:       sales,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

func (r *saleRepository) GetMetrics(ctx context.Context) (*domain.SaleMetrics, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var metrics domain.SaleMetrics

	// Ventas hoy
	r.db.WithContext(ctx).Model(&saleModel{}).
		Where("created_at >= ?", startOfDay).
		Count(&metrics.TodaySalesCount)

	var todayRevenue *float64
	r.db.WithContext(ctx).Model(&saleModel{}).
		Where("created_at >= ?", startOfDay).
		Select("SUM(total_amount)").Scan(&todayRevenue)
	if todayRevenue != nil {
		metrics.TodayTotalRevenue = *todayRevenue
	}

	// Ventas del mes
	r.db.WithContext(ctx).Model(&saleModel{}).
		Where("created_at >= ?", startOfMonth).
		Count(&metrics.MonthSalesCount)

	var monthRevenue *float64
	r.db.WithContext(ctx).Model(&saleModel{}).
		Where("created_at >= ?", startOfMonth).
		Select("SUM(total_amount)").Scan(&monthRevenue)
	if monthRevenue != nil {
		metrics.MonthTotalRevenue = *monthRevenue
	}

	if metrics.MonthSalesCount > 0 {
		metrics.AverageTicket = math.Round((metrics.MonthTotalRevenue/float64(metrics.MonthSalesCount))*100) / 100
	}

	return &metrics, nil
}

func toSaleModel(s *domain.Sale) *saleModel {
	items := make([]saleItemModel, len(s.Items))
	for i, item := range s.Items {
		items[i] = saleItemModel{
			ID:               item.ID,
			SaleID:           s.ID,
			ProductID:        item.ProductID,
			ProductName:      item.ProductName,
			ProductSKU:       item.ProductSKU,
			ProductUnit:      item.ProductUnit,
			PresentationID:   item.PresentationID,
			PresentationName: item.PresentationName,
			Quantity:         item.Quantity,
			ConversionFactor: item.ConversionFactor,
			BaseQuantity:     item.BaseQuantity,
			UnitPrice:        item.UnitPrice,
			Subtotal:         item.Subtotal,
		}
	}

	return &saleModel{
		ID:            s.ID,
		InvoiceNumber: s.InvoiceNumber,
		CustomerID:    s.CustomerID,
		CustomerName:  s.CustomerName,
		UserID:        s.UserID,
		UserName:      s.UserName,
		TotalAmount:   s.TotalAmount,
		PaymentMethod: string(s.PaymentMethod),
		AmountPaid:    s.AmountPaid,
		ChangeAmount:  s.ChangeAmount,
		Notes:         s.Notes,
		Items:         items,
		CreatedAt:     s.CreatedAt,
	}
}

func toSaleDomain(m *saleModel) *domain.Sale {
	items := make([]domain.SaleItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = domain.SaleItem{
			ID:               item.ID,
			SaleID:           item.SaleID,
			ProductID:        item.ProductID,
			ProductName:      item.ProductName,
			ProductSKU:       item.ProductSKU,
			ProductUnit:      item.ProductUnit,
			PresentationID:   item.PresentationID,
			PresentationName: item.PresentationName,
			Quantity:         item.Quantity,
			ConversionFactor: item.ConversionFactor,
			BaseQuantity:     item.BaseQuantity,
			UnitPrice:        item.UnitPrice,
			Subtotal:         item.Subtotal,
		}
	}

	return &domain.Sale{
		ID:            m.ID,
		InvoiceNumber: m.InvoiceNumber,
		CustomerID:    m.CustomerID,
		CustomerName:  m.CustomerName,
		UserID:        m.UserID,
		UserName:      m.UserName,
		TotalAmount:   m.TotalAmount,
		PaymentMethod: domain.PaymentMethod(m.PaymentMethod),
		AmountPaid:    m.AmountPaid,
		ChangeAmount:  m.ChangeAmount,
		Notes:         m.Notes,
		Items:         items,
		CreatedAt:     m.CreatedAt,
	}
}
