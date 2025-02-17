package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProduct_UpdatePrice(t *testing.T) {
	p := Product{Price: 10.0}
	p.UpdatePrice(20.0)
	assert.Equal(t, 20.0, p.Price)
}

func TestProduct_Sell(t *testing.T) {
	p := Product{Quantity: 10}
	err := p.Sell(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, p.Quantity)

	err = p.Sell(10)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientStock, err)
}

func TestProduct_Restock(t *testing.T) {
	p := Product{Quantity: 10}
	p.Restock(5)
	assert.Equal(t, 15, p.Quantity)
}

func TestProduct_Display(t *testing.T) {
	p := Product{ID: 1, Name: "Test Product", Price: 10.0, Quantity: 5, Category: "Test"}
	expected := "ID: 1, Name: Test Product, Price: 10.00, Quantity: 5, Category: Test"
	assert.Equal(t, expected, p.Display())
}

func TestInventory_AddProduct(t *testing.T) {
	i := Inventory{Products: make(map[int]Product)}
	p := Product{ID: 1, Name: "Test Product"}

	err := i.AddProduct(p)
	assert.NoError(t, err)
	assert.Equal(t, p, i.Products[1])

	err = i.AddProduct(p)
	assert.Error(t, err)
	assert.Equal(t, ErrProductAlreadyExists, err)
}

func TestInventory_RemoveProduct(t *testing.T) {
	i := Inventory{Products: make(map[int]Product)}
	p := Product{ID: 1, Name: "Test Product"}
	i.Products[1] = p

	err := i.RemoveProduct(1)
	assert.NoError(t, err)
	assert.NotContains(t, i.Products, 1)

	err = i.RemoveProduct(1)
	assert.Error(t, err)
	assert.Equal(t, ErrProductNotFound, err)
}

func TestInventory_FindProductByName(t *testing.T) {
	i := Inventory{Products: make(map[int]Product)}
	p := Product{ID: 1, Name: "Test Product"}
	i.Products[1] = p

	product, err := i.FindProductByName("Test Product")
	assert.NoError(t, err)
	assert.Equal(t, &p, product)

	product, err = i.FindProductByName("Nonexistent Product")
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, ErrProductNotFound, err)
}

func TestInventory_ListByCategory(t *testing.T) {
	i := Inventory{Products: make(map[int]Product)}
	p1 := Product{ID: 1, Name: "Product 1", Category: "Category 1"}
	p2 := Product{ID: 2, Name: "Product 2", Category: "Category 2"}
	p3 := Product{ID: 3, Name: "Product 3", Category: "Category 1"}
	i.Products[1] = p1
	i.Products[2] = p2
	i.Products[3] = p3

	products := i.ListByCategory("Category 1")
	assert.Len(t, products, 2)
	assert.Contains(t, products, p1)
	assert.Contains(t, products, p3)
}

func TestInventory_TotalValue(t *testing.T) {
	i := Inventory{Products: make(map[int]Product)}
	p1 := Product{ID: 1, Price: 10.0, Quantity: 2}
	p2 := Product{ID: 2, Price: 20.0, Quantity: 1}
	i.Products[1] = p1
	i.Products[2] = p2

	totalValue := i.TotalValue()
	assert.Equal(t, 40.0, totalValue)
}

func TestMemoryStorage_Save(t *testing.T) {
	m := MemoryStorage{Products: make(map[int]Product)}
	p := Product{ID: 1, Name: "Test Product"}

	err := m.Save(p)
	assert.NoError(t, err)
	assert.Equal(t, p, m.Products[1])
}

func TestMemoryStorage_GetByID(t *testing.T) {
	m := MemoryStorage{Products: make(map[int]Product)}
	p := Product{ID: 1, Name: "Test Product"}
	m.Products[1] = p

	product, err := m.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, &p, product)

	product, err = m.GetByID(2)
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, ErrProductNotFound, err)
}

func TestMemoryStorage_Delete(t *testing.T) {
	m := MemoryStorage{Products: make(map[int]Product)}
	p := Product{ID: 1, Name: "Test Product"}
	m.Products[1] = p

	err := m.Delete(1)
	assert.NoError(t, err)
	assert.NotContains(t, m.Products, 1)

	err = m.Delete(1)
	assert.Error(t, err)
	assert.Equal(t, ErrProductNotFound, err)
}
