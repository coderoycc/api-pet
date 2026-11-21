package application

import (
	"context"
	"fmt"
	"math"
	"time"

	auditDomain "api-go/internal/audit/domain"
	inventoryDomain "api-go/internal/inventory/domain"
	productDomain "api-go/internal/products/domain"
	"api-go/internal/purchases/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseService struct {
	repo          domain.PurchaseRepository
	productRepo   productDomain.ProductRepository
	inventoryRepo inventoryDomain.InventoryRepository
	auditService  auditDomain.AuditService
	db            *gorm.DB
}

func NewPurchaseService(
	repo domain.PurchaseRepository,
	productRepo productDomain.ProductRepository,
	inventoryRepo inventoryDomain.InventoryRepository,
	auditService auditDomain.AuditService,
	db *gorm.DB,
) *PurchaseService {
	return &PurchaseService{
		repo:          repo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		auditService:  auditService,
		db:            db,
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

	if s.auditService != nil {
		s.auditService.Record(ctx, auditDomain.AuditEntry{
			Module:      "purchases",
			Action:      auditDomain.ActionCreate,
			EntityType:  "purchase",
			EntityID:    createdPurchase.ID.String(),
			Description: fmt.Sprintf("Orden de compra %s creada", createdPurchase.Number),
			After: map[string]any{
				"number":       createdPurchase.Number,
				"status":       createdPurchase.Status,
				"total_amount": createdPurchase.TotalAmount,
			},
			Metadata: map[string]any{
				"supplier_id":  createdPurchase.SupplierID.String(),
				"warehouse_id": createdPurchase.WarehouseID.String(),
				"items_count":  len(createdPurchase.Items),
			},
		})
	}

	return toPurchaseResponse(createdPurchase), nil
}

func (s *PurchaseService) ReceivePurchase(ctx context.Context, id uuid.UUID, userID, userName string) (*PurchaseResponse, error) {
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

	now := time.Now()

	// Transacción atómica ACID
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range purchase.Items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return err
			}

			oldStock := product.Quantity
			oldCost := product.Cost
			qtyIn := float64(item.Quantity)
			unitCostIn := item.UnitCost

			newStock := oldStock + qtyIn

			// Cálculo del Costo Promedio Ponderado (CPP):
			// Si no había stock previo o era <= 0, el nuevo costo es el de esta compra.
			// De lo contrario, se pondera el valor total previo con el valor ingresado.
			var newCPP float64
			if oldStock <= 0 {
				newCPP = unitCostIn
			} else {
				totalVal := (oldStock * oldCost) + (qtyIn * unitCostIn)
				newCPP = totalVal / newStock
			}

			// Actualizar producto en unidad base
			product.Quantity = newStock
			product.Cost = math.Round(newCPP*10000) / 10000 // 4 decimales para precisión contable
			if err := s.productRepo.Update(ctx, product); err != nil {
				return err
			}

			// Crear o registrar lote para FEFO / FIFO
			batchNumber := fmt.Sprintf("LOTE-%s-%s", purchase.Number, product.SKU)
			batch := &inventoryDomain.InventoryBatch{
				ID:          uuid.New(),
				ProductID:   product.ID,
				BatchNumber: batchNumber,
				Quantity:    item.Quantity,
				Cost:        &unitCostIn,
				SupplierID:  &purchase.SupplierID,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := s.inventoryRepo.CreateBatch(ctx, batch); err != nil {
				return err
			}

			// Registrar movimiento de Entrada en Kardex (Regla de Oro de Inventarios)
			reason := fmt.Sprintf("Recepción Orden de Compra %s", purchase.Number)
			reasonType := "purchase"
			notes := fmt.Sprintf("Ingreso de mercadería por orden de compra. Proveedor: %s. Almacén: %s", purchase.SupplierName, purchase.WarehouseName)

			kardexLog := &inventoryDomain.InventoryLog{
				ID:            uuid.New(),
				ProductID:     product.ID,
				ProductName:   product.Name,
				BatchID:       &batch.ID,
				BatchNumber:   &batch.BatchNumber,
				SKU:           product.SKU,
				Type:          inventoryDomain.LogTypeInbound,
				Quantity:      item.Quantity,
				PreviousStock: int(math.Round(oldStock)),
				NewStock:      int(math.Round(newStock)),
				Reason:        reason,
				ReasonType:    &reasonType,
				UserID:        userID,
				UserName:      userName,
				Notes:         &notes,
				CreatedAt:     now,
				Undoable:      false,
			}
			if err := s.inventoryRepo.CreateLog(ctx, kardexLog); err != nil {
				return err
			}
		}

		// Actualizar estado de la Orden de Compra
		purchase.Status = domain.PurchaseStatusReceived
		purchase.ReceivedAt = &now
		purchase.UpdatedAt = now

		if err := s.repo.UpdateStatus(ctx, purchase); err != nil {
			return err
		}

		if s.auditService != nil {
			if err := s.auditService.RecordTx(ctx, tx, auditDomain.AuditEntry{
				Module:      "purchases",
				Action:      auditDomain.ActionReceive,
				EntityType:  "purchase",
				EntityID:    purchase.ID.String(),
				Description: fmt.Sprintf("Orden de compra %s recibida", purchase.Number),
				Before:      map[string]any{"status": domain.PurchaseStatusPending},
				After:       map[string]any{"status": domain.PurchaseStatusReceived},
				Metadata: map[string]any{
					"purchase_number": purchase.Number,
					"total_amount":    purchase.TotalAmount,
					"items_count":     len(purchase.Items),
				},
			}); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
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

	previousStatus := purchase.Status
	purchase.Status = domain.PurchaseStatusCancelled
	purchase.UpdatedAt = time.Now()

	if err := s.repo.UpdateStatus(ctx, purchase); err != nil {
		return nil, err
	}

	if s.auditService != nil {
		s.auditService.Record(ctx, auditDomain.AuditEntry{
			Module:      "purchases",
			Action:      auditDomain.ActionCancel,
			EntityType:  "purchase",
			EntityID:    purchase.ID.String(),
			Description: fmt.Sprintf("Orden de compra %s cancelada", purchase.Number),
			Before:      map[string]any{"status": previousStatus},
			After:       map[string]any{"status": domain.PurchaseStatusCancelled},
			Metadata: map[string]any{
				"purchase_number": purchase.Number,
			},
		})
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

	data := make([]*PurchaseResponse, 0, len(result.Data))
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
