package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
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

	httpCtx := context.Background()
	httpCtx, cancelHttpCtx := context.WithTimeout(httpCtx, time.Millisecond*300)
	defer cancelHttpCtx()

	dbCtx := context.Background()
	dbCtx, cancelDbCtx := context.WithTimeout(dbCtx, time.Millisecond*10)
	defer cancelDbCtx()

	select {
	case <-ctx.Done():
		cancelHttpCtx()
		cancelDbCtx()
		w.WriteHeader(http.StatusRequestTimeout)
		return
	default:
		if r.URL.Path != "/cotacao" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		select {
		case <-httpCtx.Done():
			w.WriteHeader(http.StatusRequestTimeout)
			return
		default:
			quotation, err := GetQuotation(httpCtx)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			select {
			case <-httpCtx.Done():
				w.WriteHeader(http.StatusRequestTimeout)
				return
			default:
				SaveQuotation(dbCtx, quotation)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(quotation.Bid)
		}
	}
}

func GetQuotation(ctx context.Context) (*Quotation, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
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

func SaveQuotation(ctx context.Context, quotation *Quotation) error {
	db, err := sql.Open("sqlite3", "./quotations.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`
		insert into quotation(id, code, code_in, name, high, low, var_bid, pct_change, bid, ask, timestamp, created_date) 
		values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		uuid.New().String(),
		quotation.Code,
		quotation.CodeIn,
		quotation.Name,
		quotation.High,
		quotation.Low,
		quotation.VarBid,
		quotation.PctChange,
		quotation.Bid,
		quotation.Ask,
		quotation.Timestamp,
		quotation.CreateDate,
	)
	if err != nil {
		return err
	}

	return nil
}
