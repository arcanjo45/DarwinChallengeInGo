package service

import (
	"encoding/json"
	"fmt"
	"log"
	"main/data"
	"main/settings"
	"net/http"
	"time"
)

const (
	layout = "2006-01-02"
)

type Result struct {
	Amount float64 `json:"amount"`
	Base   string  `json:"base"`
	Date   string  `json:"date"`
	Rates  struct {
		USD float64 `json:"USD"` // Note the lowercase "usd" tag; Go struct field names are case-sensitive
	} `json:"rates"`
}

func FetchLatestRate() float64 {
	centralBankUrl := "https://api.frankfurter.app/latest?to=USD"

	response, err := http.Get(centralBankUrl)

	if err != nil {
		log.Fatal("Failed to request rates")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", response.StatusCode)
	}

	ratesInfo := Result{}

	err = json.NewDecoder(response.Body).Decode(&ratesInfo)
	if err != nil {
		log.Fatalf("Failed to parse response body: %v", err)
	}

	date, _ := time.Parse("2006-01-02", ratesInfo.Date)

	settings.DBConnector.Create(&data.Rate{Date: date, Value: ratesInfo.Rates.USD})

	return ratesInfo.Rates.USD

}

type CurrencyRate struct {
	USD float64 `json:"USD"`
}

// Define a struct for the rates map

type Rates map[string]CurrencyRate

// Define the main struct that represents the entire payload

type Payload struct {
	Amount    float64 `json:"amount"`
	Base      string  `json:"base"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Rates     Rates   `json:"rates"`
}

func FetchIntervalsRate(StartDate string, EndDate string) map[string]CurrencyRate {

	requestUrl := fmt.Sprintf("https://api.frankfurter.app/%s..%s?to=USD", StartDate, EndDate)

	response, err := http.Get(requestUrl)

	if err != nil {
		log.Fatal("Failed to request rates")
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", response.StatusCode)
	}

	defer response.Body.Close()

	ratesInfo := Payload{}

	err = json.NewDecoder(response.Body).Decode(&ratesInfo)
	if err != nil {
		log.Fatalf("Failed to parse response body: %v", err)
	}

	fmt.Printf("%+v\n", ratesInfo)

	return ratesInfo.Rates
}

func CheckDatesInDB(StartDate string, EndDate string) []time.Time {

	startDate, _ := time.Parse(layout, StartDate)
	endDate, _ := time.Parse(layout, EndDate)

	var dbDates []time.Time

	var intervalDates []time.Time

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		intervalDates = append(intervalDates, d)
	}

	settings.DBConnector.Model(&data.Rate{}).Where("date BETWEEN ? AND ?", startDate, endDate).Pluck("date", &dbDates)

	missingDates := intersection(intervalDates, dbDates)

	fmt.Print(missingDates)

	return missingDates
}

func FetchDayRate(date time.Time) data.Rate {
	centralBankUrl := fmt.Sprintf("https://api.frankfurter.app/%s?to=USD", date.Format(layout))
	log.Print(centralBankUrl)

	response, err := http.Get(centralBankUrl)

	if err != nil {
		log.Fatal("Failed to request rates")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", response.StatusCode)
	}

	ratesInfo := Result{}

	err = json.NewDecoder(response.Body).Decode(&ratesInfo)
	if err != nil {
		log.Fatalf("Failed to parse response body: %v", err)
	}

	dayRate := data.Rate{Date: date, Value: ratesInfo.Rates.USD}

	return dayRate
}

func FetchAVGRate(StartDate time.Time, EndDate time.Time) float64 {
	var avgRate float64
	row := settings.DBConnector.Model(&data.Rate{}).Select("AVG (value)").Where("date BETWEEN ? AND ?", StartDate, EndDate).Row()

	err := row.Scan(&avgRate)

	if err != nil {
		fmt.Println("Error scanning average rate:", err)
		return 0
	}

	return avgRate
}

func intersection(first, second []time.Time) []time.Time {
	m := make(map[time.Time]bool)
	for _, item := range second {
		m[item] = true
	}

	var diff []time.Time
	for _, item := range first {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}

	return diff
}
