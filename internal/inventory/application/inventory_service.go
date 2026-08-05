package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"api-go/internal/inventory/domain"
	productDomain "api-go/internal/products/domain"

	"github.com/google/uuid"
)

type InventoryService interface {
	GetBatches(ctx context.Context, productID *uuid.UUID) ([]*domain.InventoryBatch, error)
	GetCriticalStock(ctx context.Context, minStock int) ([]*domain.CriticalStockItem, error)
	GetMetrics(ctx context.Context) (*domain.InventoryMetrics, error)
	GetLogs(ctx context.Context, filters *domain.InventoryFilters) (*domain.PaginatedLogsResult, error)
	AdjustStock(ctx context.Context, req domain.AdjustStockRequest, userID, userName string) (*domain.InventoryLog, *domain.StockAdjustment, error)
	QuickAdjust(ctx context.Context, adjustments []domain.QuickAdjustment, userID, userName string) ([]*domain.InventoryLog, error)
	RegisterInbound(ctx context.Context, op domain.InboundOperation, userID, userName string) (*domain.InventoryLog, error)
	RegisterOutbound(ctx context.Context, op domain.OutboundOperation, userID, userName string) ([]*domain.InventoryLog, error)
	RegisterSupplierReturn(ctx context.Context, op domain.SupplierReturnOperation, userID, userName string) ([]*domain.InventoryLog, error)
	UndoLog(ctx context.Context, logID uuid.UUID) (*domain.InventoryLog, error)
}

type inventoryService struct {
	repo        domain.InventoryRepository
	productRepo productDomain.ProductRepository
}

func NewInventoryService(repo domain.InventoryRepository, productRepo productDomain.ProductRepository) InventoryService {
	return &inventoryService{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *inventoryService) GetBatches(ctx context.Context, productID *uuid.UUID) ([]*domain.InventoryBatch, error) {
	if productID != nil && *productID != uuid.Nil {
		return s.repo.GetBatchesByProductID(ctx, *productID)
	}
	return s.repo.GetAllBatches(ctx)
}

func (s *inventoryService) GetCriticalStock(ctx context.Context, minStock int) ([]*domain.CriticalStockItem, error) {
	if minStock <= 0 {
		minStock = 5
	}
	return s.repo.GetCriticalStock(ctx, minStock)
}

func (s *inventoryService) GetMetrics(ctx context.Context) (*domain.InventoryMetrics, error) {
	return s.repo.GetMetrics(ctx)
}

func (s *inventoryService) GetLogs(ctx context.Context, filters *domain.InventoryFilters) (*domain.PaginatedLogsResult, error) {
	return s.repo.GetLogs(ctx, filters)
}

func (s *inventoryService) AdjustStock(ctx context.Context, req domain.AdjustStockRequest, userID, userName string) (*domain.InventoryLog, *domain.StockAdjustment, error) {
	if req.NewStock < 0 {
		return nil, nil, domain.ErrInvalidStockQuantity
	}

	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, nil, err
	}

	var targetBatch *domain.InventoryBatch
	if req.BatchID != nil && *req.BatchID != uuid.Nil {
		targetBatch, err = s.repo.GetBatchByID(ctx, *req.BatchID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		batches, err := s.repo.GetBatchesByProductID(ctx, req.ProductID)
		if err == nil && len(batches) > 0 {
			targetBatch = batches[0]
		}
	}

	oldStock := 0
	if targetBatch != nil {
		oldStock = targetBatch.Quantity
	}

	diffQuantity := req.NewStock - oldStock

	now := time.Now()
	var batchID *uuid.UUID
	var batchNumber *string

	if targetBatch != nil {
		targetBatch.Quantity = req.NewStock
		targetBatch.UpdatedAt = now
		if err := s.repo.UpdateBatchQuantity(ctx, targetBatch.ID, req.NewStock); err != nil {
			return nil, nil, err
		}
		batchID = &targetBatch.ID
		batchNumber = &targetBatch.BatchNumber
	}

	// Update overall product quantity
	product.Quantity += diffQuantity
	if product.Quantity < 0 {
		product.Quantity = 0
	}
	_ = s.productRepo.Update(ctx, product)

	log := &domain.InventoryLog{
		ID:            uuid.New(),
		ProductID:     product.ID,
		ProductName:   product.Name,
		BatchID:       batchID,
		BatchNumber:   batchNumber,
		SKU:           product.SKU,
		Type:          domain.LogTypeAdjustment,
		Quantity:      diffQuantity,
		PreviousStock: oldStock,
		NewStock:      req.NewStock,
		Reason:        req.Reason,
		UserID:        userID,
		UserName:      userName,
		CreatedAt:     now,
		Undoable:      true,
	}

	if err := s.repo.CreateLog(ctx, log); err != nil {
		return nil, nil, err
	}

	adjustment := &domain.StockAdjustment{
		ID:          uuid.New(),
		ProductID:   product.ID,
		ProductName: product.Name,
		BatchID:     batchID,
		BatchNumber: batchNumber,
		SKU:         product.SKU,
		OldStock:    oldStock,
		NewStock:    req.NewStock,
		Reason:      req.Reason,
		UserID:      userID,
		UserName:    userName,
		CreatedAt:   now,
	}

	if err := s.repo.CreateStockAdjustment(ctx, adjustment); err != nil {
		return nil, nil, err
	}

	return log, adjustment, nil
}

func (s *inventoryService) QuickAdjust(ctx context.Context, adjustments []domain.QuickAdjustment, userID, userName string) ([]*domain.InventoryLog, error) {
	var logs []*domain.InventoryLog
	for _, adj := range adjustments {
		req := domain.AdjustStockRequest{
			ProductID: adj.ProductID,
			BatchID:   adj.BatchID,
			NewStock:  adj.NewStock,
			Reason:    adj.Reason,
		}
		if req.Reason == "" {
			req.Reason = "Ajuste rápido"
		}
		log, _, err := s.AdjustStock(ctx, req, userID, userName)
		if err == nil && log != nil {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func (s *inventoryService) RegisterInbound(ctx context.Context, op domain.InboundOperation, userID, userName string) (*domain.InventoryLog, error) {
	if op.Quantity <= 0 {
		return nil, domain.ErrInvalidStockQuantity
	}

	product, err := s.productRepo.GetByID(ctx, op.ProductID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	batchNum := fmt.Sprintf("LOTE-%s", uuid.New().String()[:6])
	if op.BatchNumber != nil && *op.BatchNumber != "" {
		batchNum = *op.BatchNumber
	}

	cost := product.Cost
	if op.Cost != nil {
		cost = *op.Cost
	}

	newBatch := &domain.InventoryBatch{
		ID:             uuid.New(),
		ProductID:      product.ID,
		BatchNumber:    batchNum,
		ExpirationDate: op.ExpirationDate,
		Quantity:       op.Quantity,
		Cost:           &cost,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateBatch(ctx, newBatch); err != nil {
		return nil, err
	}

	// Update product stock
	product.Quantity += op.Quantity
	_ = s.productRepo.Update(ctx, product)

	reason := "Entrada de mercadería"
	if op.Reason != "" {
		reason = op.Reason
	}

	log := &domain.InventoryLog{
		ID:            uuid.New(),
		ProductID:     product.ID,
		ProductName:   product.Name,
		BatchID:       &newBatch.ID,
		BatchNumber:   &newBatch.BatchNumber,
		SKU:           product.SKU,
		Type:          domain.LogTypeInbound,
		Quantity:      op.Quantity,
		PreviousStock: 0,
		NewStock:      op.Quantity,
		Reason:        reason,
		ReasonType:    op.ReasonType,
		UserID:        userID,
		UserName:      userName,
		Notes:         op.Notes,
		CreatedAt:     now,
		Undoable:      true,
	}

	if err := s.repo.CreateLog(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *inventoryService) RegisterOutbound(ctx context.Context, op domain.OutboundOperation, userID, userName string) ([]*domain.InventoryLog, error) {
	if op.Quantity <= 0 {
		return nil, domain.ErrInvalidStockQuantity
	}

	product, err := s.productRepo.GetByID(ctx, op.ProductID)
	if err != nil {
		return nil, err
	}

	availableBatches, err := s.repo.GetAvailableBatchesForProduct(ctx, op.ProductID)
	if err != nil {
		return nil, err
	}

	totalAvailable := 0
	for _, b := range availableBatches {
		totalAvailable += b.Quantity
	}

	if op.Quantity > totalAvailable {
		return nil, domain.ErrInsufficientStock
	}

	// Order batches FEFO (First Expire First Out) or FIFO
	sort.Slice(availableBatches, func(i, j int) bool {
		if availableBatches[i].ExpirationDate != nil && availableBatches[j].ExpirationDate != nil {
			return availableBatches[i].ExpirationDate.Before(*availableBatches[j].ExpirationDate)
		}
		return availableBatches[i].CreatedAt.Before(availableBatches[j].CreatedAt)
	})

	remaining := op.Quantity
	var logs []*domain.InventoryLog
	now := time.Now()

	reason := "Salida de mercadería"
	if op.Reason != "" {
		reason = op.Reason
	}

	for _, batch := range availableBatches {
		if remaining <= 0 {
			break
		}

		toRemove := batch.Quantity
		if remaining < toRemove {
			toRemove = remaining
		}

		prevStock := batch.Quantity
		newStock := batch.Quantity - toRemove
		batch.Quantity = newStock

		if err := s.repo.UpdateBatchQuantity(ctx, batch.ID, newStock); err != nil {
			return nil, err
		}

		remaining -= toRemove

		log := &domain.InventoryLog{
			ID:            uuid.New(),
			ProductID:     product.ID,
			ProductName:   product.Name,
			BatchID:       &batch.ID,
			BatchNumber:   &batch.BatchNumber,
			SKU:           product.SKU,
			Type:          domain.LogTypeOutbound,
			Quantity:      toRemove,
			PreviousStock: prevStock,
			NewStock:      newStock,
			Reason:        reason,
			ReasonType:    op.ReasonType,
			UserID:        userID,
			UserName:      userName,
			Notes:         op.Notes,
			CreatedAt:     now,
			Undoable:      true,
		}

		if err := s.repo.CreateLog(ctx, log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	// Update product stock
	product.Quantity -= op.Quantity
	if product.Quantity < 0 {
		product.Quantity = 0
	}
	_ = s.productRepo.Update(ctx, product)

	return logs, nil
}

func (s *inventoryService) RegisterSupplierReturn(ctx context.Context, op domain.SupplierReturnOperation, userID, userName string) ([]*domain.InventoryLog, error) {
	if op.Quantity <= 0 {
		return nil, domain.ErrInvalidStockQuantity
	}

	product, err := s.productRepo.GetByID(ctx, op.ProductID)
	if err != nil {
		return nil, err
	}

	var targetBatches []*domain.InventoryBatch
	if op.BatchID != nil && *op.BatchID != uuid.Nil {
		b, err := s.repo.GetBatchByID(ctx, *op.BatchID)
		if err != nil {
			return nil, err
		}
		targetBatches = []*domain.InventoryBatch{b}
	} else {
		targetBatches, err = s.repo.GetAvailableBatchesForProduct(ctx, op.ProductID)
		if err != nil {
			return nil, err
		}
	}

	totalAvailable := 0
	for _, b := range targetBatches {
		totalAvailable += b.Quantity
	}

	if op.Quantity > totalAvailable {
		return nil, domain.ErrInsufficientStock
	}

	remaining := op.Quantity
	var logs []*domain.InventoryLog
	now := time.Now()

	formattedNotes := fmt.Sprintf("Tipo de devolución: %s. ", op.ReturnType)
	if op.RefundAmount != nil {
		formattedNotes += fmt.Sprintf("Monto esperado: $%.2f. ", *op.RefundAmount)
	}
	if op.Notes != nil && *op.Notes != "" {
		formattedNotes += fmt.Sprintf("Notas: %s", *op.Notes)
	}

	reason := "Devolución a proveedor"
	if op.Reason != "" {
		reason = op.Reason
	}
	reasonType := op.ReasonType

	for _, batch := range targetBatches {
		if remaining <= 0 {
			break
		}

		toRemove := batch.Quantity
		if remaining < toRemove {
			toRemove = remaining
		}

		prevStock := batch.Quantity
		newStock := batch.Quantity - toRemove
		batch.Quantity = newStock

		if err := s.repo.UpdateBatchQuantity(ctx, batch.ID, newStock); err != nil {
			return nil, err
		}

		remaining -= toRemove

		log := &domain.InventoryLog{
			ID:            uuid.New(),
			ProductID:     product.ID,
			ProductName:   product.Name,
			BatchID:       &batch.ID,
			BatchNumber:   &batch.BatchNumber,
			SKU:           product.SKU,
			Type:          domain.LogTypeSupplierReturn,
			Quantity:      toRemove,
			PreviousStock: prevStock,
			NewStock:      newStock,
			Reason:        reason,
			ReasonType:    &reasonType,
			UserID:        userID,
			UserName:      userName,
			Notes:         &formattedNotes,
			CreatedAt:     now,
			Undoable:      true,
		}

		if err := s.repo.CreateLog(ctx, log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	// Update product stock
	product.Quantity -= op.Quantity
	if product.Quantity < 0 {
		product.Quantity = 0
	}
	_ = s.productRepo.Update(ctx, product)

	return logs, nil
}

func (s *inventoryService) UndoLog(ctx context.Context, logID uuid.UUID) (*domain.InventoryLog, error) {
	originalLog, err := s.repo.GetLogByID(ctx, logID)
	if err != nil {
		return nil, err
	}
	if !originalLog.Undoable {
		return nil, domain.ErrLogNotUndoable
	}

	product, err := s.productRepo.GetByID(ctx, originalLog.ProductID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	quantityChange := 0

	if originalLog.BatchID != nil {
		batch, err := s.repo.GetBatchByID(ctx, *originalLog.BatchID)
		if err == nil && batch != nil {
			if originalLog.Type == domain.LogTypeInbound {
				batch.Quantity -= originalLog.Quantity
				if batch.Quantity < 0 {
					batch.Quantity = 0
				}
				quantityChange = -originalLog.Quantity
			} else if originalLog.Type == domain.LogTypeOutbound || originalLog.Type == domain.LogTypeSupplierReturn {
				batch.Quantity += originalLog.Quantity
				quantityChange = originalLog.Quantity
			} else if originalLog.Type == domain.LogTypeAdjustment {
				quantityChange = originalLog.PreviousStock - originalLog.NewStock
				batch.Quantity = originalLog.PreviousStock
			}
			_ = s.repo.UpdateBatchQuantity(ctx, batch.ID, batch.Quantity)
		}
	}

	// Update product overall stock
	product.Quantity += quantityChange
	if product.Quantity < 0 {
		product.Quantity = 0
	}
	_ = s.productRepo.Update(ctx, product)

	// Mark original log as no longer undoable
	_ = s.repo.UpdateLogUndoable(ctx, originalLog.ID, false)

	undoLog := &domain.InventoryLog{
		ID:            uuid.New(),
		ProductID:     originalLog.ProductID,
		ProductName:   originalLog.ProductName,
		BatchID:       originalLog.BatchID,
		BatchNumber:   originalLog.BatchNumber,
		SKU:           originalLog.SKU,
		Type:          domain.LogTypeAdjustment,
		Quantity:      quantityChange,
		PreviousStock: originalLog.NewStock,
		NewStock:      originalLog.NewStock + quantityChange,
		Reason:        fmt.Sprintf("Deshacer: %s", originalLog.Reason),
		UserID:        originalLog.UserID,
		UserName:      originalLog.UserName,
		CreatedAt:     now,
		Undoable:      false,
	}

	if err := s.repo.CreateLog(ctx, undoLog); err != nil {
		return nil, err
	}

	return undoLog, nil
}
