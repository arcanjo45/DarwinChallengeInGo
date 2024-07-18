package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"main/data"
	"main/service"
	"main/settings"
	"net/http"
	"time"
)

const (
	layout = "2006-01-02"
)

func main() {
	err := settings.InitDB()
	router := gin.Default()
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	migrateErr := settings.DBConnector.AutoMigrate(data.Rate{})
	if migrateErr != nil {
		return
	}

	router.GET("/rates", AvgRateHandler)
	router.GET("/rates/latest", LatestRateHandler)

	router.Run(":9999")

}

func LatestRateHandler(context *gin.Context) {
	latestRate := service.FetchLatestRate()
	response := map[string]interface{}{
		"rate": latestRate,
	}
	context.JSON(http.StatusOK, response)
}

func AvgRateHandler(context *gin.Context) {
	startDate := context.Query("startDate")
	endDate := context.Query("endDate")

	missingInfo := service.CheckDatesInDB(startDate, endDate)

	var newRates []data.Rate

	if missingInfo != nil {
		for _, date := range missingInfo {
			newRates = append(newRates, service.FetchDayRate(date))
		}
		settings.DBConnector.CreateInBatches(&newRates, 10)
	}

	startDateTime, _ := time.Parse(layout, startDate)
	endDateTime, _ := time.Parse(layout, endDate)

	avgRate := service.FetchAVGRate(startDateTime, endDateTime)
	fmt.Printf("Current USD rate: %f", avgRate)

	response := map[string]interface{}{
		"rate": avgRate,
	}
	context.JSON(http.StatusOK, response)
}
