package routes

import (
	"github.com/Sparsh1401/car_listing_service/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, carController *controllers.CarController) {
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/cars", carController.GetCars)
			v1.GET("/cars/:id", carController.GetCarByID)
			v1.POST("/cars", carController.CreateCar)
			v1.PUT("/cars/:id", carController.UpdateCar)
			v1.DELETE("/cars/:id", carController.DeleteCar)
			v1.POST("/scrape", carController.ScrapeCars)
		}
	}
}
