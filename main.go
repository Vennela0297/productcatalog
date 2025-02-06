package main

import (
	"net/http"
	"strconv"

	productcatalog "repo/product"

	"github.com/gin-gonic/gin"
)

var inventory = productcatalog.Inventory{Products: make(map[int]productcatalog.Product)}

func main() {
	// Initialize Gin router
	router := gin.Default()

	// Define CRUD routes
	router.POST("/products", createProduct)
	router.GET("/products/:id", getProduct)
	router.PUT("/products/:id", updateProduct)
	router.DELETE("/products/:id", deleteProduct)
	router.GET("/products", listProducts)

	// Run the server
	router.Run(":8080")
}

func createProduct(c *gin.Context) {
	var product productcatalog.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try to add the product to the inventory
	if err := inventory.AddProduct(product); err != nil {
		if err == productcatalog.ErrProductAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "product already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, product)
}

func getProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	product, exists := inventory.Products[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func updateProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	var updatedProduct productcatalog.Product
	if err := c.ShouldBindJSON(&updatedProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	product, exists := inventory.Products[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	product.Name = updatedProduct.Name
	product.Price = updatedProduct.Price
	product.Quantity = updatedProduct.Quantity
	product.Category = updatedProduct.Category
	inventory.Products[id] = product
	c.JSON(http.StatusOK, product)
}

func deleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	_, exists := inventory.Products[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	delete(inventory.Products, id)
	c.JSON(http.StatusNoContent, nil)
}

func listProducts(c *gin.Context) {
	products := make([]productcatalog.Product, 0, len(inventory.Products))
	for _, product := range inventory.Products {
		products = append(products, product)
	}
	c.JSON(http.StatusOK, products)
}
