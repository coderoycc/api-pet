package postgres

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"api-go/internal/products/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productModel struct {
	ID          uuid.UUID           `gorm:"type:uuid;primaryKey"`
	SKU         string              `gorm:"type:varchar(100);not null;uniqueIndex"`
	Name        string              `gorm:"type:varchar(255);not null"`
	Description string              `gorm:"type:text"`
	Category    string              `gorm:"type:varchar(100)"`
	Subcategory string              `gorm:"type:varchar(100)"`
	Price       float64             `gorm:"type:decimal(10,2);not null"`
	Cost        float64             `gorm:"type:decimal(10,2);not null"`
	Quantity    int                 `gorm:"not null"`
	MinQuantity int                 `gorm:"not null"`
	MaxQuantity *int                
	Unit        string              `gorm:"type:varchar(50)"`
	Status      string              `gorm:"type:varchar(50)"`
	SupplierID  *uuid.UUID          `gorm:"type:uuid"`
	Tags        []string            `gorm:"type:jsonb;serializer:json"`
	Images      []productImageModel `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Batches     []productBatchModel `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt   time.Time           `gorm:"not null"`
	UpdatedAt   time.Time           `gorm:"not null"`
}

func (productModel) TableName() string {
	return "products"
}

type productImageModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
	URL       string    `gorm:"type:varchar(500);not null"`
	Alt       string    `gorm:"type:varchar(255)"`
	IsPrimary bool      `gorm:"default:false"`
}

func (productImageModel) TableName() string {
	return "product_images"
}

type productBatchModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProductID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	BatchNumber    string     `gorm:"type:varchar(100);not null"`
	ExpirationDate time.Time  `gorm:"type:date;not null"`
	Quantity       int        `gorm:"not null"`
	SupplierID     *uuid.UUID `gorm:"type:uuid"`
}

func (productBatchModel) TableName() string {
	return "product_batches"
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&productModel{}, &productImageModel{}, &productBatchModel{})
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	model := toProductModel(product)
	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateSKU
		}
		return err
	}
	return nil
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var model productModel
	err := r.db.WithContext(ctx).Preload("Images").Preload("Batches").First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}
	return toProductDomain(&model), nil
}

func (r *productRepository) GetAll(ctx context.Context, page, limit int, filters *domain.ProductFilters, sortBy, sortOrder string) (*domain.PaginatedResult, error) {
	var models []productModel
	var total int64

	query := r.db.WithContext(ctx).Model(&productModel{})

	if filters != nil {
		if filters.Search != "" {
			search := "%" + strings.ToLower(filters.Search) + "%"
			query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(sku) LIKE ?", search, search, search)
		}
		if filters.Category != "" {
			query = query.Where("LOWER(category) = ?", strings.ToLower(filters.Category))
		}
		if filters.Status != "" {
			query = query.Where("status = ?", filters.Status)
		}
		if filters.MinPrice != nil {
			query = query.Where("price >= ?", *filters.MinPrice)
		}
		if filters.MaxPrice != nil {
			query = query.Where("price <= ?", *filters.MaxPrice)
		}
		if filters.SupplierID != nil {
			query = query.Where("supplier_id = ?", *filters.SupplierID)
		}
		if len(filters.Tags) > 0 {
			// jsonb ?& array[...] for multiple tags, or check sequentially for simpler driver compatibility.
			for _, tag := range filters.Tags {
				query = query.Where("tags @> ?", `"`+tag+`"`) // simple JSONB containment check for string array
			}
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
		query = query.Order(sortBy + " " + order)
	} else {
		query = query.Order("created_at DESC")
	}

	offset := (page - 1) * limit
	err = query.Preload("Images").Offset(offset).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}

	products := make([]*domain.Product, len(models))
	for i, m := range models {
		products[i] = toProductDomain(&m)
	}

	return &domain.PaginatedResult{
		Data:       products,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	model := toProductModel(product)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&productModel{}).Where("id = ?", product.ID).Updates(map[string]interface{}{
			"sku":          model.SKU,
			"name":         model.Name,
			"description":  model.Description,
			"category":     model.Category,
			"subcategory":  model.Subcategory,
			"price":        model.Price,
			"cost":         model.Cost,
			"quantity":     model.Quantity,
			"min_quantity": model.MinQuantity,
			"max_quantity": model.MaxQuantity,
			"unit":         model.Unit,
			"status":       model.Status,
			"supplier_id":  model.SupplierID,
			"tags":         model.Tags,
			"updated_at":   time.Now(),
		}).Error
		if err != nil {
			if isDuplicateKey(err) {
				return domain.ErrDuplicateSKU
			}
			return err
		}

		// Reemplazar imágenes por simplicidad o manejar individualmente
		if len(model.Images) > 0 {
			if err := tx.Where("product_id = ?", product.ID).Delete(&productImageModel{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.Images).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&productModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *productRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).Model(&productModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *productRepository) GetProductsByCategory(ctx context.Context, category string) ([]*domain.Product, error) {
	var models []productModel
	err := r.db.WithContext(ctx).Preload("Images").Where("LOWER(category) = ?", strings.ToLower(category)).Find(&models).Error
	if err != nil {
		return nil, err
	}

	products := make([]*domain.Product, len(models))
	for i, m := range models {
		products[i] = toProductDomain(&m)
	}
	return products, nil
}

func (r *productRepository) GetExpiringProducts(ctx context.Context, filters *domain.ExpiringProductFilters) ([]*domain.ExpiringProduct, int64, error) {
	now := time.Now()
	daysAhead := 30
	if filters != nil && filters.DaysAhead != nil {
		daysAhead = *filters.DaysAhead
	}
	cutoff := now.AddDate(0, 0, daysAhead)

	query := r.db.WithContext(ctx).Model(&productBatchModel{}).
		Preload("Product"). // Assuming we need product details, wait, we don't have Product relation setup in struct yet, let's just do a join.
		Joins("JOIN products ON products.id = product_batches.product_id").
		Joins("LEFT JOIN suppliers ON suppliers.id = products.supplier_id").
		Select("product_batches.*, products.name as product_name, products.sku as product_sku, products.category as product_category, products.unit as product_unit, suppliers.name as supplier_name").
		Where("product_batches.expiration_date > ? AND product_batches.expiration_date <= ?", now, cutoff)

	// Filtering
	if filters != nil {
		if filters.Search != "" {
			search := "%" + strings.ToLower(filters.Search) + "%"
			query = query.Where("LOWER(products.name) LIKE ? OR LOWER(products.sku) LIKE ? OR LOWER(product_batches.batch_number) LIKE ?", search, search, search)
		}
		if filters.Category != "" {
			query = query.Where("LOWER(products.category) = ?", strings.ToLower(filters.Category))
		}
	}

	type resultStruct struct {
		productBatchModel
		ProductName      string
		ProductSKU       string
		ProductCategory  string
		ProductUnit      string
		SupplierName     string
	}

	var results []resultStruct
	err := query.Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	var expiringProducts []*domain.ExpiringProduct
	for _, r := range results {
		daysUntil := int(math.Ceil(r.ExpirationDate.Sub(now).Hours() / 24))
		urgency := "normal"
		if daysUntil <= 7 {
			urgency = "critical"
		} else if daysUntil <= 15 {
			urgency = "warning"
		}

		if filters != nil && filters.Urgency != "" && filters.Urgency != urgency {
			continue // Skip if doesn't match requested urgency
		}

		expiringProducts = append(expiringProducts, &domain.ExpiringProduct{
			ProductID:       r.ProductID,
			ProductName:     r.ProductName,
			SKU:             r.ProductSKU,
			Category:        r.ProductCategory,
			BatchID:         r.ID,
			BatchNumber:     r.BatchNumber,
			ExpirationDate:  r.ExpirationDate.Format("2006-01-02"),
			DaysUntilExpiry: daysUntil,
			Quantity:        r.Quantity,
			Unit:            r.ProductUnit,
			Urgency:         urgency,
			SupplierName:    r.SupplierName,
		})
	}

	return expiringProducts, int64(len(expiringProducts)), nil
}

func (r *productRepository) GetLowStockProducts(ctx context.Context, threshold int) ([]*domain.Product, error) {
	var models []productModel
	err := r.db.WithContext(ctx).Preload("Images").Where("quantity <= min_quantity OR quantity <= ?", threshold).Find(&models).Error
	if err != nil {
		return nil, err
	}

	products := make([]*domain.Product, len(models))
	for i, m := range models {
		products[i] = toProductDomain(&m)
	}
	return products, nil
}

// Helpers

func isDuplicateKey(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505"))
}

func toProductDomain(m *productModel) *domain.Product {
	p := &domain.Product{
		ID:          m.ID,
		SKU:         m.SKU,
		Name:        m.Name,
		Description: m.Description,
		Category:    m.Category,
		Subcategory: m.Subcategory,
		Price:       m.Price,
		Cost:        m.Cost,
		Quantity:    m.Quantity,
		MinQuantity: m.MinQuantity,
		MaxQuantity: m.MaxQuantity,
		Unit:        m.Unit,
		Status:      m.Status,
		SupplierID:  m.SupplierID,
		Tags:        m.Tags,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	
	if len(m.Images) > 0 {
		p.Images = make([]domain.ProductImage, len(m.Images))
		for i, img := range m.Images {
			p.Images[i] = domain.ProductImage{
				ID:        img.ID,
				ProductID: img.ProductID,
				URL:       img.URL,
				Alt:       img.Alt,
				IsPrimary: img.IsPrimary,
			}
		}
	}

	if len(m.Batches) > 0 {
		p.Batches = make([]domain.ProductBatch, len(m.Batches))
		for i, batch := range m.Batches {
			p.Batches[i] = domain.ProductBatch{
				ID:             batch.ID,
				ProductID:      batch.ProductID,
				BatchNumber:    batch.BatchNumber,
				ExpirationDate: batch.ExpirationDate,
				Quantity:       batch.Quantity,
				SupplierID:     batch.SupplierID,
			}
		}
	}

	return p
}

func toProductModel(p *domain.Product) *productModel {
	m := &productModel{
		ID:          p.ID,
		SKU:         p.SKU,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Subcategory: p.Subcategory,
		Price:       p.Price,
		Cost:        p.Cost,
		Quantity:    p.Quantity,
		MinQuantity: p.MinQuantity,
		MaxQuantity: p.MaxQuantity,
		Unit:        p.Unit,
		Status:      p.Status,
		SupplierID:  p.SupplierID,
		Tags:        p.Tags,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	
	if len(p.Images) > 0 {
		m.Images = make([]productImageModel, len(p.Images))
		for i, img := range p.Images {
			m.Images[i] = productImageModel{
				ID:        img.ID,
				ProductID: img.ProductID,
				URL:       img.URL,
				Alt:       img.Alt,
				IsPrimary: img.IsPrimary,
			}
		}
	}

	if len(p.Batches) > 0 {
		m.Batches = make([]productBatchModel, len(p.Batches))
		for i, batch := range p.Batches {
			m.Batches[i] = productBatchModel{
				ID:             batch.ID,
				ProductID:      batch.ProductID,
				BatchNumber:    batch.BatchNumber,
				ExpirationDate: batch.ExpirationDate,
				Quantity:       batch.Quantity,
				SupplierID:     batch.SupplierID,
			}
		}
	}

	return m
}
