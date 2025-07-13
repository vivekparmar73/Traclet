package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
}

type Category struct {
	gorm.Model
	Name string `json:"name" gorm:"unique"`
	Type string `json:"type"`
}

type Transaction struct {
	gorm.Model
	UserID      uint    `json:"user_id"`
	CategoryID  uint    `json:"category_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

var DB *gorm.DB

func main() {
	dbUser := "root"
	dbPassword := "1234"
	dbName := "expense_tracker"

	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(&User{}, &Category{}, &Transaction{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	initialCategories := []Category{
		{Name: "Salary", Type: "income"},
		{Name: "Food", Type: "expense"},
		{Name: "Transport", Type: "expense"},
		{Name: "Entertainment", Type: "expense"},
	}

	for _, category := range initialCategories {
		DB.FirstOrCreate(&category, Category{Name: category.Name})
	}

	r := gin.Default()

	// Recovery middleware to handle panics
	r.Use(gin.Recovery())

	// Routes
	r.POST("/register", registerUser)
	r.POST("/login", loginUser)
	r.GET("/categories", getCategories)

	transactionRoutes := r.Group("/transactions")
	transactionRoutes.Use(basicAuthMiddleware())
	{
		transactionRoutes.POST("/", createTransaction)
		transactionRoutes.GET("/", getTransactions)
		transactionRoutes.GET("/summary", getSummary)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func basicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("User-ID") // Changed from "User_id" to "User-ID"
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User-ID header required"})
			c.Abort()
			return
		}

		var user User
		if err := DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func registerUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.Name == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, email and password are required"})
		return
	}

	var existingUser User
	if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created", "id": user.ID})
}

func loginUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := DB.Where("email = ? AND password = ?", input.Email, input.Password).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user_id": user.ID})
}

func getCategories(c *gin.Context) {
	var categories []Category
	if err := DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func createTransaction(c *gin.Context) {
	userID := c.GetString("userID")

	var input struct {
		CategoryID  uint    `json:"category_id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		Date        string  `json:"date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category Category
	if err := DB.First(&category, input.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	transaction := Transaction{
		UserID:      parseUint(userID),
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		Description: input.Description,
		Date:        input.Date,
	}

	if err := DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create transaction"})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

func getTransactions(c *gin.Context) {
	userID := c.GetString("userID")

	var transactions []Transaction
	if err := DB.Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func getSummary(c *gin.Context) {
	userID := c.GetString("userID")
	month := c.Query("month")
	year := c.Query("year")

	if month == "" || year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Month and year parameters required"})
		return
	}

	startDate := fmt.Sprintf("%s-%s-01", year, month)
	endDate := fmt.Sprintf("%s-%s-31", year, month)

	var transactions []Transaction
	if err := DB.Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch transactions"})
		return
	}

	var totalIncome, totalExpense float64
	categorySummary := make(map[string]float64)

	for _, t := range transactions {
		var category Category
		DB.First(&category, t.CategoryID)

		if category.Type == "income" {
			totalIncome += t.Amount
		} else {
			totalExpense += t.Amount
		}
		categorySummary[category.Name] += t.Amount
	}

	c.JSON(http.StatusOK, gin.H{
		"total_income":     totalIncome,
		"total_expense":    totalExpense,
		"net_balance":      totalIncome - totalExpense,
		"category_summary": categorySummary, // Fixed typo from "category_summary" to "category_summary"
	})
}

func parseUint(s string) uint {
	var i uint
	fmt.Sscanf(s, "%d", &i)
	return i
}
