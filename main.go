package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	productcatalog "repo/product"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type OrderCreated struct {
	OrderID   int    `json:"order_id"`
	ProductID int    `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}

var (
	db          *gorm.DB
	kafkaWriter *kafka.Writer
	inventory   = productcatalog.Inventory{Products: make(map[int]productcatalog.Product)}
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Initialize PostgreSQL connection
	dsn := "host=" + os.Getenv("DB_HOST") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" port=" + os.Getenv("DB_PORT") +
		" sslmode=disable TimeZone=Asia/Shanghai"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	// Auto migrate the Product model
	db.AutoMigrate(&productcatalog.Product{})

	// Initialize Kafka writer
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(os.Getenv("KAFKA_BROKER")),
		Topic:    "product-events",
		Balancer: &kafka.LeastBytes{},
	}

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

	// Check if a product with the same ID already exists
	var existingProduct productcatalog.Product
	if err := db.Where("id = ?", product.ID).First(&existingProduct).Error; err == nil {
		// Product with the same ID already exists
		c.JSON(http.StatusConflict, gin.H{"error": "Product with the same ID already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create the new product
	if err := db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
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

	// Update the product in the database
	if err := db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish event to Kafka
	orderCreated := OrderCreated{
		OrderID:   id,
		ProductID: id,
		Quantity:  product.Quantity,
		Status:    "updated",
	}
	orderCreatedJSON, err := json.Marshal(orderCreated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize order event"})
		return
	}

	err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Value: orderCreatedJSON,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish order event"})
		return
	}

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

	// Delete the product from the database
	if err := db.Delete(&productcatalog.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish event to Kafka
	orderCreated := OrderCreated{
		OrderID:   id,
		ProductID: id,
		Quantity:  0,
		Status:    "deleted",
	}
	orderCreatedJSON, err := json.Marshal(orderCreated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize order event"})
		return
	}

	err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Value: orderCreatedJSON,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish order event"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func listProducts(c *gin.Context) {
	products := make([]productcatalog.Product, 0, len(inventory.Products))
	for _, product := range inventory.Products {
		products = append(products, product)
	}
	c.JSON(http.StatusOK, products)
}
