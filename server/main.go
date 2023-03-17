package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Quotation struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type QuotationWrapper struct {
	USDBRL Quotation `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", GetQuotationHandler)
	http.ListenAndServe(":8080", nil)
}

func GetQuotationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	select {
	case <-ctx.Done():
		w.WriteHeader(http.StatusRequestTimeout)
		return
	default:
		if r.URL.Path != "/cotacao" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		quotation, err := GetQuotation()
		if err != nil {
			panic(err)
		}

		err = SaveQuotation(quotation)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(quotation.Bid)
	}
}

func GetQuotation() (*Quotation, error) {
	httpCtx := context.Background()
	httpCtx, cancelHttpCtx := context.WithTimeout(httpCtx, time.Millisecond*200)
	defer cancelHttpCtx()

	req, err := http.NewRequestWithContext(httpCtx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var quotation QuotationWrapper
	err = json.Unmarshal(body, &quotation)
	if err != nil {
		return nil, err
	}
	return &quotation.USDBRL, nil
}

func SaveQuotation(quotation *Quotation) error {
	dbCtx := context.Background()
	dbCtx, cancelDbCtx := context.WithTimeout(dbCtx, time.Nanosecond*10)
	defer cancelDbCtx()

	db, err := gorm.Open(sqlite.Open("quotations.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Quotation{})

	err = db.WithContext(dbCtx).Create(&quotation).Error
	if err != nil {
		return err
	}

	return nil
}
