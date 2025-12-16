package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/yourusername/car-listing-service/models"
	"github.com/yourusername/car-listing-service/services"
	"github.com/gin-gonic/gin"
)

type CarController struct {
	service services.CarService
}

func NewCarController(service services.CarService) *CarController {
	return &CarController{service: service}
}

func (ctrl *CarController) GetCars(c *gin.Context) {
	cars, err := ctrl.service.GetAllCars()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cars)
}

func (ctrl *CarController) GetCarByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	car, err := ctrl.service.GetCarByID(int(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if car == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
		return
	}

	c.JSON(http.StatusOK, car)
}

func (ctrl *CarController) CreateCar(c *gin.Context) {
	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.service.CreateCar(&car); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, car)
}

func (ctrl *CarController) UpdateCar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	car.ID = int(id)
	if err := ctrl.service.UpdateCar(&car); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, car)
}

func (ctrl *CarController) DeleteCar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	if err := ctrl.service.DeleteCar(int(id)); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Car deleted successfully"})
}

func (ctrl *CarController) ScrapeCars(c *gin.Context) {
	count, err := ctrl.service.ScrapeAndStoreCars()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Scraping finished with error: " + err.Error(),
			"count": count,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Scraping completed successfully",
		"count":   count,
	})
}
