package domain

import "errors"

var (
	ErrBatchNotFound        = errors.New("lote de inventario no encontrado")
	ErrLogNotFound          = errors.New("registro de inventario no encontrado")
	ErrInsufficientStock    = errors.New("stock insuficiente para realizar la operación")
	ErrInvalidStockQuantity = errors.New("el stock no puede ser negativo")
	ErrLogNotUndoable       = errors.New("este registro no se puede deshacer o ya fue deshecho")
)
