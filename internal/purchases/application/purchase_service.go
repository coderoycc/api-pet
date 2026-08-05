package application

import (
	"context"
	"time"

	productDomain "api-go/internal/products/domain"
	"api-go/internal/purchases/domain"

	"github.com/google/uuid"
)

type PurchaseService struct {
	repo        domain.PurchaseRepository
	productRepo productDomain.ProductRepository
}

func NewPurchaseService(repo domain.PurchaseRepository, productRepo productDomain.ProductRepository) *PurchaseService {
	return &PurchaseService{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *PurchaseService) CreatePurchase(ctx context.Context, dto *PurchaseCreateDto) (*PurchaseResponse, error) {
	supplierID, err := uuid.Parse(dto.SupplierID)
	if err != nil {
		return nil, err
	}

	warehouseID, err := uuid.Parse(dto.WarehouseID)
	if err != nil {
		return nil, err
	}

	var items []domain.PurchaseItem
	var totalAmount float64

	for _, itemDto := range dto.Items {
		prodID, err := uuid.Parse(itemDto.ProductID)
		if err != nil {
			return nil, err
		}

		// Valida que el producto exista
		_, err = s.productRepo.GetByID(ctx, prodID)
		if err != nil {
			return nil, err
		}

		subtotal := float64(itemDto.Quantity) * itemDto.UnitCost
		totalAmount += subtotal

		items = append(items, domain.PurchaseItem{
			ID:        uuid.New(),
			ProductID: prodID,
			Quantity:  itemDto.Quantity,
			UnitCost:  itemDto.UnitCost,
			Subtotal:  subtotal,
		})
	}

	purchase := &domain.Purchase{
		ID:          uuid.New(),
		SupplierID:  supplierID,
		WarehouseID: warehouseID,
		Items:       items,
		Status:      domain.PurchaseStatusPending,
		TotalAmount: totalAmount,
		Notes:       dto.Notes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// El número de orden "Number" será generado y asignado dentro del repositorio usando una secuencia de BD

	if err := s.repo.Create(ctx, purchase); err != nil {
		return nil, err
	}

	// Obtenemos la compra creada para incluir los nombres (Supplier, Warehouse, Products) generados por los Joins
	createdPurchase, err := s.repo.GetByID(ctx, purchase.ID)
	if err != nil {
		return nil, err
	}

	return toPurchaseResponse(createdPurchase), nil
}

func (s *PurchaseService) ReceivePurchase(ctx context.Context, id uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if purchase.Status == domain.PurchaseStatusCancelled {
		return nil, domain.ErrCannotReceiveCancelled
	}
	if purchase.Status == domain.PurchaseStatusReceived {
		return nil, domain.ErrAlreadyReceived
	}

	// Update stock for each product
	for _, item := range purchase.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err == nil {
			product.Quantity += item.Quantity
			_ = s.productRepo.Update(ctx, product) // Asumiendo update básico. Podría requerir una transacción mayor en producción.
		}
	}

	now := time.Now()
	purchase.Status = domain.PurchaseStatusReceived
	purchase.ReceivedAt = &now
	purchase.UpdatedAt = now

	if err := s.repo.UpdateStatus(ctx, purchase); err != nil {
		return nil, err
	}

	return toPurchaseResponse(purchase), nil
}

func (s *PurchaseService) CancelPurchase(ctx context.Context, id uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if purchase.Status == domain.PurchaseStatusReceived {
		return nil, domain.ErrCannotCancelReceived
	}

	purchase.Status = domain.PurchaseStatusCancelled
	purchase.UpdatedAt = time.Now()

	if err := s.repo.UpdateStatus(ctx, purchase); err != nil {
		return nil, err
	}

	return toPurchaseResponse(purchase), nil
}

func (s *PurchaseService) GetPurchaseByID(ctx context.Context, id uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPurchaseResponse(purchase), nil
}

func (s *PurchaseService) GetPurchases(ctx context.Context, page, limit int, filters *domain.PurchaseFilters, sortBy, sortOrder string) (*PurchaseListResponse, error) {
	result, err := s.repo.GetAll(ctx, page, limit, filters, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	var data []*PurchaseResponse
	for _, p := range result.Data {
		data = append(data, toPurchaseResponse(p))
	}

	return &PurchaseListResponse{
		Data:       data,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}, nil
}

func (s *PurchaseService) GetMetrics(ctx context.Context) (*PurchaseMetricsResponse, error) {
	metrics, err := s.repo.GetMetrics(ctx)
	if err != nil {
		return nil, err
	}

	return &PurchaseMetricsResponse{
		TotalOrders:          metrics.TotalOrders,
		PendingOrders:        metrics.PendingOrders,
		ReceivedThisMonth:    metrics.ReceivedThisMonth,
		TotalAmountThisMonth: metrics.TotalAmountThisMonth,
	}, nil
}

func toPurchaseResponse(purchase *domain.Purchase) *PurchaseResponse {
	resp := &PurchaseResponse{
		ID:            purchase.ID.String(),
		Number:        purchase.Number,
		SupplierID:    purchase.SupplierID.String(),
		SupplierName:  purchase.SupplierName,
		WarehouseID:   purchase.WarehouseID.String(),
		WarehouseName: purchase.WarehouseName,
		Status:        purchase.Status,
		TotalAmount:   purchase.TotalAmount,
		Notes:         purchase.Notes,
		CreatedAt:     purchase.CreatedAt,
		ReceivedAt:    purchase.ReceivedAt,
		UpdatedAt:     purchase.UpdatedAt,
	}

	var items []PurchaseItemResponse
	for _, item := range purchase.Items {
		items = append(items, PurchaseItemResponse{
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitCost:    item.UnitCost,
			Unit:        item.Unit,
			Subtotal:    item.Subtotal,
		})
	}
	resp.Items = items

	return resp
}
