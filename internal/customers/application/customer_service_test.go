package application_test

import (
	"context"
	"testing"
	"time"

	"api-go/internal/customers/application"
	"api-go/internal/customers/domain"

	"github.com/google/uuid"
)

// --- Mock Repository ---

type mockCustomerRepo struct {
	customers        []*domain.Customer
	cities           []string
	conflictResult   bool
	associatedResult bool
	err              error
}

func (m *mockCustomerRepo) Create(ctx context.Context, customer *domain.Customer) error {
	if m.err != nil {
		return m.err
	}
	m.customers = append(m.customers, customer)
	return nil
}

func (m *mockCustomerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, c := range m.customers {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, domain.ErrCustomerNotFound
}

func (m *mockCustomerRepo) GetAll(ctx context.Context, page, limit int, filters *domain.CustomerFilters, sortBy string, descending bool) (*domain.PaginatedCustomerResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	total := int64(len(m.customers))
	return &domain.PaginatedCustomerResult{
		Data:       m.customers,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}, nil
}

func (m *mockCustomerRepo) Search(ctx context.Context, query string, limit int, status string) ([]*domain.Customer, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.customers, nil
}

func (m *mockCustomerRepo) GetDistinctCities(ctx context.Context) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cities, nil
}

func (m *mockCustomerRepo) Update(ctx context.Context, customer *domain.Customer) error {
	return m.err
}

func (m *mockCustomerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	for i, c := range m.customers {
		if c.ID == id {
			m.customers = append(m.customers[:i], m.customers[i+1:]...)
			return nil
		}
	}
	return domain.ErrCustomerNotFound
}

func (m *mockCustomerRepo) CheckDocumentConflict(ctx context.Context, docType, docNumber string, excludeID *uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.conflictResult, nil
}

func (m *mockCustomerRepo) HasAssociatedRecords(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.associatedResult, nil
}

// --- Helper ---

func newValidCreateDTO() application.CustomerCreateDTO {
	email := "juan@test.com"
	return application.CustomerCreateDTO{
		Name:           "Juan Pérez",
		DocumentType:   "ci",
		DocumentNumber: "4567891",
		Email:          &email,
		Status:         "active",
	}
}

// --- Tests ---

func TestCreateCustomer_Success(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	customer, err := svc.CreateCustomer(context.Background(), newValidCreateDTO())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if customer.Name != "Juan Pérez" {
		t.Errorf("expected name 'Juan Pérez', got %s", customer.Name)
	}
	if customer.Status != "active" {
		t.Errorf("expected status 'active', got %s", customer.Status)
	}
	if customer.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestCreateCustomer_DefaultStatus(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	dto := newValidCreateDTO()
	dto.Status = "" // Sin status

	customer, err := svc.CreateCustomer(context.Background(), dto)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if customer.Status != "active" {
		t.Errorf("expected default status 'active', got %s", customer.Status)
	}
}

func TestCreateCustomer_MissingName(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	dto := newValidCreateDTO()
	dto.Name = ""

	_, err := svc.CreateCustomer(context.Background(), dto)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestCreateCustomer_InvalidDocumentType(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	dto := newValidCreateDTO()
	dto.DocumentType = "invalid_type"

	_, err := svc.CreateCustomer(context.Background(), dto)
	if err == nil {
		t.Fatal("expected error for invalid document type, got nil")
	}
}

func TestCreateCustomer_InvalidEmail(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	dto := newValidCreateDTO()
	badEmail := "not-an-email"
	dto.Email = &badEmail

	_, err := svc.CreateCustomer(context.Background(), dto)
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
}

func TestCreateCustomer_DocumentConflict(t *testing.T) {
	repo := &mockCustomerRepo{conflictResult: true}
	svc := application.NewCustomerService(repo)

	_, err := svc.CreateCustomer(context.Background(), newValidCreateDTO())
	if err == nil {
		t.Fatal("expected ErrCustomerDocumentExists, got nil")
	}
}

func TestGetCustomers_Pagination(t *testing.T) {
	repo := &mockCustomerRepo{
		customers: []*domain.Customer{
			{ID: uuid.New(), Name: "A", DocumentType: "ci", DocumentNumber: "1", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New(), Name: "B", DocumentType: "nit", DocumentNumber: "2", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	svc := application.NewCustomerService(repo)

	res, err := svc.GetCustomers(context.Background(), 1, 10, nil, "createdAt", false)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if res.Pagination.CurrentPage != 1 {
		t.Errorf("expected currentPage 1, got %d", res.Pagination.CurrentPage)
	}
	if res.Pagination.PageSize != 10 {
		t.Errorf("expected pageSize 10, got %d", res.Pagination.PageSize)
	}
	if res.Pagination.TotalItems != 2 {
		t.Errorf("expected totalItems 2, got %d", res.Pagination.TotalItems)
	}
	if res.Total != 2 {
		t.Errorf("expected total 2 (dual), got %d", res.Total)
	}
}

func TestGetCustomers_PageNormalization(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	// page=0 y limit=200 deben normalizarse
	res, err := svc.GetCustomers(context.Background(), 0, 200, nil, "", false)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if res.Pagination.CurrentPage != 1 {
		t.Errorf("expected page normalized to 1, got %d", res.Pagination.CurrentPage)
	}
	if res.Pagination.PageSize != 100 {
		t.Errorf("expected limit capped at 100, got %d", res.Pagination.PageSize)
	}
}

func TestSearchCustomers_ShortQuery(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	results, err := svc.SearchCustomers(context.Background(), "a", 10, "active")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results for query < 2 chars, got %d", len(results))
	}
}

func TestSearchCustomers_ValidQuery(t *testing.T) {
	repo := &mockCustomerRepo{
		customers: []*domain.Customer{
			{ID: uuid.New(), Name: "Juan Pérez", DocumentType: "ci", DocumentNumber: "123", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	svc := application.NewCustomerService(repo)

	results, err := svc.SearchCustomers(context.Background(), "ju", 10, "active")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Juan Pérez" {
		t.Errorf("expected name 'Juan Pérez', got %s", results[0].Name)
	}
}

func TestGetDistinctCities(t *testing.T) {
	repo := &mockCustomerRepo{
		cities: []string{"Cochabamba", "La Paz", "Santa Cruz"},
	}
	svc := application.NewCustomerService(repo)

	cities, err := svc.GetDistinctCities(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cities) != 3 {
		t.Errorf("expected 3 cities, got %d", len(cities))
	}
}

func TestDeleteCustomer_WithAssociatedRecords(t *testing.T) {
	id := uuid.New()
	repo := &mockCustomerRepo{
		customers:        []*domain.Customer{{ID: id, Name: "Test", DocumentType: "ci", DocumentNumber: "1", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		associatedResult: true,
	}
	svc := application.NewCustomerService(repo)

	err := svc.DeleteCustomer(context.Background(), id)
	if err == nil {
		t.Fatal("expected ErrCustomerHasAssociatedRecords, got nil")
	}
}

func TestDeleteCustomer_NotFound(t *testing.T) {
	repo := &mockCustomerRepo{}
	svc := application.NewCustomerService(repo)

	err := svc.DeleteCustomer(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent customer, got nil")
	}
}

func TestUpdateCustomer_InvalidDocumentType(t *testing.T) {
	id := uuid.New()
	repo := &mockCustomerRepo{
		customers: []*domain.Customer{
			{ID: id, Name: "Test", DocumentType: "ci", DocumentNumber: "1", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	svc := application.NewCustomerService(repo)

	dto := application.CustomerUpdateDTO{
		Name:           "Updated",
		DocumentType:   "invalid",
		DocumentNumber: "1",
	}

	_, err := svc.UpdateCustomer(context.Background(), id, dto)
	if err == nil {
		t.Fatal("expected error for invalid document type, got nil")
	}
}
