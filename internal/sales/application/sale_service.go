package application

import (
	"context"
	"fmt"
	"math"
	"time"

	auditDomain "api-go/internal/audit/domain"
	inventoryDomain "api-go/internal/inventory/domain"
	productDomain "api-go/internal/products/domain"
	"api-go/internal/sales/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SaleService struct {
	repo          domain.SaleRepository
	productRepo   productDomain.ProductRepository
	inventoryRepo inventoryDomain.InventoryRepository
	auditService  auditDomain.AuditService
	db            *gorm.DB
}

func NewSaleService(
	repo domain.SaleRepository,
	productRepo productDomain.ProductRepository,
	inventoryRepo inventoryDomain.InventoryRepository,
	auditService auditDomain.AuditService,
	db *gorm.DB,
) *SaleService {
	return &SaleService{
		repo:          repo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		auditService:  auditService,
		db:            db,
	}
}

func (s *SaleService) ProcessSale(ctx context.Context, dto *SaleCreateDto, userID uuid.UUID, userName string) (*SaleResponse, error) {
	if len(dto.Items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	now := time.Now()
	var totalAmount float64
	var saleItems []domain.SaleItem

	// Preparar y validar ítems
	for _, itemDto := range dto.Items {
		prodID, err := uuid.Parse(itemDto.ProductID)
		if err != nil {
			return nil, fmt.Errorf("ID de producto inválido: %s", itemDto.ProductID)
		}

		if itemDto.Quantity <= 0 {
			return nil, fmt.Errorf("la cantidad debe ser mayor a 0 para el producto %s", itemDto.ProductID)
		}

		product, err := s.productRepo.GetByID(ctx, prodID)
		if err != nil {
			return nil, fmt.Errorf("producto no encontrado: %w", err)
		}

		factor := itemDto.ConversionFactor
		if factor <= 0 {
			factor = 1.0
		}

		baseQty := itemDto.Quantity * factor
		if product.Quantity < baseQty {
			return nil, fmt.Errorf("stock insuficiente para '%s'. Disponible: %.2f %s, Solicitado: %.2f %s",
				product.Name, product.Quantity, product.Unit, baseQty, product.Unit)
		}

		unitPrice := itemDto.UnitPrice
		if unitPrice <= 0 {
			unitPrice = product.Price
		}

		subtotal := math.Round((itemDto.Quantity*unitPrice)*100) / 100
		totalAmount += subtotal

		var presID *uuid.UUID
		if itemDto.PresentationID != nil && *itemDto.PresentationID != "" {
			if parsed, err := uuid.Parse(*itemDto.PresentationID); err == nil {
				presID = &parsed
			}
		}

		presName := itemDto.PresentationName
		if presName == "" {
			presName = product.Unit
		}

		saleItems = append(saleItems, domain.SaleItem{
			ID:               uuid.New(),
			ProductID:        prodID,
			ProductName:      product.Name,
			ProductSKU:       product.SKU,
			ProductUnit:      product.Unit,
			PresentationID:   presID,
			PresentationName: presName,
			Quantity:         itemDto.Quantity,
			ConversionFactor: factor,
			BaseQuantity:     baseQty,
			UnitPrice:        unitPrice,
			Subtotal:         subtotal,
		})
	}

	totalAmount = math.Round(totalAmount*100) / 100

	amountPaid := dto.AmountPaid
	if amountPaid <= 0 || dto.PaymentMethod != "cash" {
		amountPaid = totalAmount
	}

	if amountPaid < totalAmount {
		return nil, domain.ErrInvalidPayment
	}

	changeAmount := math.Round((amountPaid-totalAmount)*100) / 100

	var customerID *uuid.UUID
	customerName := dto.CustomerName
	if customerName == "" {
		customerName = "Cliente Mostrador / Ocasional"
	}
	if dto.CustomerID != nil && *dto.CustomerID != "" {
		if parsed, err := uuid.Parse(*dto.CustomerID); err == nil {
			customerID = &parsed
		}
	}

	sale := &domain.Sale{
		ID:            uuid.New(),
		CustomerID:    customerID,
		CustomerName:  customerName,
		UserID:        userID,
		UserName:      userName,
		TotalAmount:   totalAmount,
		PaymentMethod: domain.PaymentMethod(dto.PaymentMethod),
		AmountPaid:    amountPaid,
		ChangeAmount:  changeAmount,
		Notes:         dto.Notes,
		Items:         saleItems,
		CreatedAt:     now,
	}

	// Ejecutar transacción ACID de venta y descuento de inventario
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Guardar la venta y generar número correlativo
		if err := s.repo.Create(ctx, sale); err != nil {
			return err
		}

		// 2. Descontar inventario físico, lotes y asentar Kardex
		for _, item := range sale.Items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return err
			}

			if product.Quantity < item.BaseQuantity {
				return domain.ErrInsufficientStock
			}

			oldStock := product.Quantity
			newStock := oldStock - item.BaseQuantity
			if newStock < 0 {
				newStock = 0
			}

			product.Quantity = newStock
			if err := s.productRepo.Update(ctx, product); err != nil {
				return err
			}

			// Descontar de lotes por FEFO
			batches, _ := s.inventoryRepo.GetAvailableBatchesForProduct(ctx, product.ID)
			remainingToDeduct := int(math.Ceil(item.BaseQuantity))
			var lastBatchID *uuid.UUID
			var lastBatchNumber *string

			for _, b := range batches {
				if remainingToDeduct <= 0 {
					break
				}
				toTake := b.Quantity
				if remainingToDeduct < toTake {
					toTake = remainingToDeduct
				}
				b.Quantity -= toTake
				_ = s.inventoryRepo.UpdateBatchQuantity(ctx, b.ID, b.Quantity)
				remainingToDeduct -= toTake
				lastBatchID = &b.ID
				lastBatchNumber = &b.BatchNumber
			}

			// Asentar movimiento en Kardex
			reason := fmt.Sprintf("Venta en mostrador Factura %s", sale.InvoiceNumber)
			reasonType := "sale"
			notes := fmt.Sprintf("Venta de %.2f %s a cliente %s", item.Quantity, item.PresentationName, sale.CustomerName)

			kardexLog := &inventoryDomain.InventoryLog{
				ID:            uuid.New(),
				ProductID:     product.ID,
				ProductName:   product.Name,
				BatchID:       lastBatchID,
				BatchNumber:   lastBatchNumber,
				SKU:           product.SKU,
				Type:          inventoryDomain.LogTypeOutbound,
				Quantity:      int(math.Ceil(item.BaseQuantity)),
				PreviousStock: int(math.Round(oldStock)),
				NewStock:      int(math.Round(newStock)),
				Reason:        reason,
				ReasonType:    &reasonType,
				UserID:        userID.String(),
				UserName:      userName,
				Notes:         &notes,
				CreatedAt:     now,
				Undoable:      true,
			}

			if err := s.inventoryRepo.CreateLog(ctx, kardexLog); err != nil {
				return err
			}
		}

		if s.auditService != nil {
			if err := s.auditService.RecordTx(ctx, tx, auditDomain.AuditEntry{
				Module:      "sales",
				Action:      auditDomain.ActionCreate,
				EntityType:  "sale",
				EntityID:    sale.ID.String(),
				Description: fmt.Sprintf("Venta %s registrada por %.2f", sale.InvoiceNumber, sale.TotalAmount),
				After: map[string]any{
					"invoice_number": sale.InvoiceNumber,
					"total_amount":   sale.TotalAmount,
					"payment_method": sale.PaymentMethod,
				},
				Metadata: map[string]any{
					"customer_name": sale.CustomerName,
					"items_count":   len(sale.Items),
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

	return toSaleResponse(sale), nil
}

func (s *SaleService) GetSaleByID(ctx context.Context, id uuid.UUID) (*SaleResponse, error) {
	sale, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toSaleResponse(sale), nil
}

func (s *SaleService) GetSales(ctx context.Context, filters *domain.SaleFilters) (*SaleListResponse, error) {
	result, err := s.repo.GetAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	dtos := make([]*SaleResponse, len(result.Data))
	for i, sale := range result.Data {
		dtos[i] = toSaleResponse(sale)
	}

	return &SaleListResponse{
		Data:       dtos,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}, nil
}

func (s *SaleService) GetMetrics(ctx context.Context) (*SaleMetricsResponse, error) {
	m, err := s.repo.GetMetrics(ctx)
	if err != nil {
		return nil, err
	}

	return &SaleMetricsResponse{
		TodaySalesCount:   m.TodaySalesCount,
		TodayTotalRevenue: m.TodayTotalRevenue,
		MonthSalesCount:   m.MonthSalesCount,
		MonthTotalRevenue: m.MonthTotalRevenue,
		AverageTicket:     m.AverageTicket,
	}, nil
}

func toSaleResponse(s *domain.Sale) *SaleResponse {
	items := make([]SaleItemResponse, len(s.Items))
	for i, item := range s.Items {
		var presID *string
		if item.PresentationID != nil {
			str := item.PresentationID.String()
			presID = &str
		}
		items[i] = SaleItemResponse{
			ID:               item.ID.String(),
			ProductID:        item.ProductID.String(),
			ProductName:      item.ProductName,
			ProductSKU:       item.ProductSKU,
			ProductUnit:      item.ProductUnit,
			PresentationID:   presID,
			PresentationName: item.PresentationName,
			Quantity:         item.Quantity,
			ConversionFactor: item.ConversionFactor,
			BaseQuantity:     item.BaseQuantity,
			UnitPrice:        item.UnitPrice,
			Subtotal:         item.Subtotal,
		}
	}

	var customerID *string
	if s.CustomerID != nil {
		str := s.CustomerID.String()
		customerID = &str
	}

	return &SaleResponse{
		ID:            s.ID.String(),
		InvoiceNumber: s.InvoiceNumber,
		CustomerID:    customerID,
		CustomerName:  s.CustomerName,
		UserID:        s.UserID.String(),
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
