package product

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrInsufficientStock           = errors.New("insufficient stock")
	ErrProductNotFound             = errors.New("product not found")
	ErrProductAlreadyExists        = errors.New("product already exists")
	ErrAPIRequestFailed            = errors.New("API request failed")
	ErrFailedToSaveProduct         = errors.New("failed to save product")
	ErrFailedToGetProduct          = errors.New("failed to get product")
	ErrFailedToDeleteProduct       = errors.New("failed to delete product")
	ErrFailedToFetchProductDetails = errors.New("failed to fetch product details")
)

// ProductStorage interface - Defines the methods for saving, retrieving, and deleting products.
type ProductStorage interface {
	Save(p Product) error
	GetByID(id int) (*Product, error)
	Delete(id int) error
}

// Product struct
type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

// Inventory struct, it will manage a collection of products.
type Inventory struct {
	Products map[int]Product
}

// MemoryStorage struct - uses a map to store products in memory.
type MemoryStorage struct {
	sync.Mutex
	Products map[int]Product
}

// NewMemoryStorage - Constructor for MemoryStorage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Products: make(map[int]Product),
	}
}

// ExternalAPI struct
type ExternalAPI struct {
	BaseURL string
}

// MockDatabaseStorage struct - simulates database interactions with random delays.
type MockDatabaseStorage struct {
	sync.Mutex
	store map[int]Product
}

// NewMockDatabaseStorage - Constructor for MockDatabaseStorage
func NewMockDatabaseStorage() *MockDatabaseStorage {
	return &MockDatabaseStorage{
		store: make(map[int]Product),
	}
}

// UpdatePrice method - Updates the product’s price.
func (p *Product) UpdatePrice(newPrice float64) {
	p.Price = newPrice
}

// Sell method - Reduces the quantity if stock is available.
func (p *Product) Sell(quantity int) error {
	if p.Quantity < quantity {
		return ErrInsufficientStock
	}
	p.Quantity -= quantity
	return nil
}

// Restock method - Increases the quantity of a product.
func (p *Product) Restock(quantity int) {
	p.Quantity += quantity
}

// Display method - Returns a formatted string with product details.
func (p *Product) Display() string {
	return fmt.Sprintf("ID: %d, Name: %s, Price: %.2f, Quantity: %d, Category: %s", p.ID, p.Name, p.Price, p.Quantity, p.Category)
}

// AddProduct method - Adds a new product to the inventory.
func (i *Inventory) AddProduct(p Product) error {
	if _, ok := i.Products[p.ID]; ok {
		return ErrProductAlreadyExists
	}
	i.Products[p.ID] = p
	return nil
}

// RemoveProduct method - Removes a product by ID.
func (i *Inventory) RemoveProduct(id int) error {
	if _, ok := i.Products[id]; !ok {
		return ErrProductNotFound
	}
	delete(i.Products, id)
	return nil
}

// FindProductByName method - Searches for a product by name.
func (i *Inventory) FindProductByName(name string) (*Product, error) {
	for _, p := range i.Products {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, ErrProductNotFound
}

// ListByCategory method - Returns all products in a given category.
func (i *Inventory) ListByCategory(category string) []Product {
	var products []Product
	for _, p := range i.Products {
		if p.Category == category {
			products = append(products, p)
		}
	}
	return products
}

// TotalValue method - Computes the total value of all products in inventory.
func (i *Inventory) TotalValue() float64 {
	var total float64
	for _, p := range i.Products {
		total += p.Price * float64(p.Quantity)
	}
	return total
}

// Save method - Saves a product to memory.
func (m *MemoryStorage) Save(p Product) error {
	m.Lock()
	defer m.Unlock()
	m.Products[p.ID] = p
	return nil
}

// GetByID method - Retrieves a product by ID from memory.
func (m *MemoryStorage) GetByID(id int) (*Product, error) {
	m.Lock()
	defer m.Unlock()
	if p, ok := m.Products[id]; ok {
		return &p, nil
	}
	return nil, ErrProductNotFound
}

// Delete method - Deletes a product by ID from memory.
func (m *MemoryStorage) Delete(id int) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.Products[id]; !ok {
		return ErrProductNotFound
	}
	delete(m.Products, id)
	return nil
}

// Save method - Saves a product to the mock database.
func (mds *MockDatabaseStorage) Save(p Product) error {
	mds.Lock()
	defer mds.Unlock()
	// Simulate database save with random delay
	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	if rand.Float32() < 0.1 {
		return ErrFailedToSaveProduct
	}
	mds.store[p.ID] = p
	return nil
}

// GetByID method - Retrieves a product by ID from the mock database.
func (mds *MockDatabaseStorage) GetByID(id int) (*Product, error) {
	mds.Lock()
	defer mds.Unlock()
	// Simulate database fetch with random delay
	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	if rand.Float32() < 0.1 {
		return nil, ErrFailedToGetProduct
	}
	if p, exists := mds.store[id]; exists {
		return &p, nil
	}
	return nil, ErrProductNotFound
}

// Delete method - Deletes a product by ID from the mock database.
func (mds *MockDatabaseStorage) Delete(id int) error {
	mds.Lock()
	defer mds.Unlock()
	// Simulate database delete with random delay
	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	if rand.Float32() < 0.1 {
		return ErrFailedToDeleteProduct
	}
	if _, exists := mds.store[id]; !exists {
		return ErrProductNotFound
	}
	delete(mds.store, id)
	return nil
}

// FetchProductDetails method - Simulates an API call to fetch product details.
func (api *ExternalAPI) FetchProductDetails(id int) (*Product, error) {
	// Simulate API request with random delay
	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
	if rand.Float32() < 0.1 {
		return nil, ErrFailedToFetchProductDetails
	}

	// Simulate fetching product details
	return &Product{
		ID:       id,
		Name:     fmt.Sprintf("Product %d", id),
		Price:    float64(rand.Intn(100)),
		Quantity: rand.Intn(100),
		Category: "Category",
	}, nil
}

// FetchProductsConcurrently function - Fetches multiple products concurrently using goroutines.
func FetchProductsConcurrently(api *ExternalAPI, ids []int) []*Product {
	var wg sync.WaitGroup
	results := make(chan *Product, len(ids))
	errors := make(chan error, len(ids))

	for _, id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			product, err := api.FetchProductDetails(id)
			if err != nil {
				errors <- err
				return
			}
			results <- product
		}(id)
	}

	wg.Wait()
	close(results)
	close(errors)

	var products []*Product
	for product := range results {
		products = append(products, product)
	}

	return products
}
