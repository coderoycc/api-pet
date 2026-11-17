package application

import (
	"context"
	"time"

	"api-go/internal/products/domain"

	"github.com/google/uuid"
)

type ProductService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, dto *ProductCreateDto) (*ProductResponse, error) {
	product := &domain.Product{
		ID:          uuid.New(),
		SKU:         dto.SKU,
		Name:        dto.Name,
		Description: dto.Description,
		Category:    dto.Category,
		Subcategory: dto.Subcategory,
		Price:       dto.Price,
		Cost:        dto.Cost,
		Quantity:    dto.Quantity,
		MinQuantity: dto.MinQuantity,
		MaxQuantity: dto.MaxQuantity,
		Unit:        dto.Unit,
		Status:      dto.Status,
		Tags:        dto.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if dto.SupplierID != nil && *dto.SupplierID != "" {
		supID, err := uuid.Parse(*dto.SupplierID)
		if err == nil {
			product.SupplierID = &supID
		}
	}

	if len(dto.Images) > 0 {
		var images []domain.ProductImage
		for _, img := range dto.Images {
			images = append(images, domain.ProductImage{
				ID:        uuid.New(),
				ProductID: product.ID,
				URL:       img.URL,
				Alt:       img.Alt,
				IsPrimary: img.IsPrimary,
			})
		}
		product.Images = images
	}

	var presentations []domain.ProductPresentation
	if len(dto.Presentations) > 0 {
		for _, p := range dto.Presentations {
			pid := uuid.New()
			if p.ID != nil && *p.ID != "" {
				if parsed, err := uuid.Parse(*p.ID); err == nil {
					pid = parsed
				}
			}
			factor := p.ConversionFactor
			if factor <= 0 {
				factor = 1.0
			}
			salePrice := p.SalePrice
			if salePrice <= 0 {
				salePrice = dto.Price
			}
			presentations = append(presentations, domain.ProductPresentation{
				ID:               pid,
				ProductID:        product.ID,
				Name:             p.Name,
				Barcode:          p.Barcode,
				ConversionFactor: factor,
				SalePrice:        salePrice,
				IsDefault:        p.IsDefault,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			})
		}
	} else {
		// Presentación por defecto con Factor 1.0
		unitName := dto.Unit
		if unitName == "" {
			unitName = "Unidad"
		}
		presentations = append(presentations, domain.ProductPresentation{
			ID:               uuid.New(),
			ProductID:        product.ID,
			Name:             unitName,
			Barcode:          dto.SKU,
			ConversionFactor: 1.0,
			SalePrice:        dto.Price,
			IsDefault:        true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})
	}
	product.Presentations = presentations

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return toProductResponse(product), nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, dto *ProductUpdateDto) (*ProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.SKU != nil {
		product.SKU = *dto.SKU
	}
	if dto.Name != nil {
		product.Name = *dto.Name
	}
	if dto.Description != nil {
		product.Description = *dto.Description
	}
	if dto.Category != nil {
		product.Category = *dto.Category
	}
	if dto.Subcategory != nil {
		product.Subcategory = *dto.Subcategory
	}
	if dto.Price != nil {
		product.Price = *dto.Price
	}
	if dto.Cost != nil {
		product.Cost = *dto.Cost
	}
	if dto.Quantity != nil {
		product.Quantity = *dto.Quantity
	}
	if dto.MinQuantity != nil {
		product.MinQuantity = *dto.MinQuantity
	}
	if dto.MaxQuantity != nil {
		product.MaxQuantity = dto.MaxQuantity
	}
	if dto.Unit != nil {
		product.Unit = *dto.Unit
	}
	if dto.Status != nil {
		product.Status = *dto.Status
	}
	if dto.Tags != nil {
		product.Tags = dto.Tags
	}
	if dto.SupplierID != nil {
		if *dto.SupplierID == "" {
			product.SupplierID = nil
		} else {
			supID, parseErr := uuid.Parse(*dto.SupplierID)
			if parseErr == nil {
				product.SupplierID = &supID
			}
		}
	}

	if dto.Images != nil {
		var images []domain.ProductImage
		for _, img := range dto.Images {
			images = append(images, domain.ProductImage{
				ID:        uuid.New(),
				ProductID: product.ID,
				URL:       img.URL,
				Alt:       img.Alt,
				IsPrimary: img.IsPrimary,
			})
		}
		product.Images = images
	}

	if dto.Presentations != nil {
		var presentations []domain.ProductPresentation
		for _, p := range dto.Presentations {
			factor := p.ConversionFactor
			if factor <= 0 {
				factor = 1.0
			}
			salePrice := p.SalePrice
			if salePrice <= 0 {
				salePrice = product.Price
			}
			presID := uuid.New()
			if p.ID != nil && *p.ID != "" {
				if parsed, err := uuid.Parse(*p.ID); err == nil {
					presID = parsed
				}
			}
			presentations = append(presentations, domain.ProductPresentation{
				ID:               presID,
				ProductID:        product.ID,
				Name:             p.Name,
				Barcode:          p.Barcode,
				ConversionFactor: factor,
				SalePrice:        salePrice,
				IsDefault:        p.IsDefault,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			})
		}
		product.Presentations = presentations
	}

	product.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	// Fetch again to get updated relations if needed, or just return the modified object
	updatedProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toProductResponse(updatedProduct), nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*ProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toProductResponse(product), nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) GetProducts(ctx context.Context, page, limit int, filters *domain.ProductFilters, sortBy, sortOrder string) (*ProductListResponse, error) {
	result, err := s.repo.GetAll(ctx, page, limit, filters, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	data := make([]*ProductResponse, 0, len(result.Data))
	for _, p := range result.Data {
		data = append(data, toProductResponse(p))
	}

	return &ProductListResponse{
		Data:       data,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}, nil
}

func (s *ProductService) UpdateProductStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *ProductService) GetExpiringProducts(ctx context.Context, filters *domain.ExpiringProductFilters) (*ExpiringProductsResponse, error) {
	data, total, err := s.repo.GetExpiringProducts(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &ExpiringProductsResponse{
		Data:  data,
		Total: total,
	}, nil
}

func (s *ProductService) GetLowStockProducts(ctx context.Context, threshold int) (*ProductListResponse, error) {
	data, err := s.repo.GetLowStockProducts(ctx, threshold)
	if err != nil {
		return nil, err
	}

	dtos := make([]*ProductResponse, 0, len(data))
	for _, p := range data {
		dtos = append(dtos, toProductResponse(p))
	}

	return &ProductListResponse{
		Data:       dtos,
		Total:      int64(len(dtos)),
		Page:       1,
		Limit:      len(dtos),
		TotalPages: 1,
	}, nil
}

func toProductResponse(product *domain.Product) *ProductResponse {
	resp := &ProductResponse{
		ID:          product.ID.String(),
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Subcategory: product.Subcategory,
		Price:       product.Price,
		Cost:        product.Cost,
		Quantity:    product.Quantity,
		MinQuantity: product.MinQuantity,
		MaxQuantity: product.MaxQuantity,
		Unit:        product.Unit,
		Status:      product.Status,
		Tags:        product.Tags,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	if product.SupplierID != nil {
		sid := product.SupplierID.String()
		resp.SupplierID = &sid
	}

	if len(product.Images) > 0 {
		var imgDtos []ProductImageDto
		for _, img := range product.Images {
			imgDtos = append(imgDtos, ProductImageDto{
				ID:        img.ID.String(),
				URL:       img.URL,
				Alt:       img.Alt,
				IsPrimary: img.IsPrimary,
			})
		}
		resp.Images = imgDtos
	} else {
		resp.Images = []ProductImageDto{}
	}

	if len(product.Presentations) > 0 {
		var presDtos []ProductPresentationDto
		for _, pres := range product.Presentations {
			presDtos = append(presDtos, ProductPresentationDto{
				ID:               pres.ID.String(),
				ProductID:        pres.ProductID.String(),
				Name:             pres.Name,
				Barcode:          pres.Barcode,
				ConversionFactor: pres.ConversionFactor,
				SalePrice:        pres.SalePrice,
				IsDefault:        pres.IsDefault,
				CreatedAt:        pres.CreatedAt,
				UpdatedAt:        pres.UpdatedAt,
			})
		}
		resp.Presentations = presDtos
	} else {
		resp.Presentations = []ProductPresentationDto{}
	}
	
	if resp.Tags == nil {
		resp.Tags = []string{}
	}

	return resp
}
