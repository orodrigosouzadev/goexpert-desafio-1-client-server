package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	SaveQuotation("./cotacao.txt", string(body))
}

func SaveQuotation(filename string, bid string) {
	s := "Dólar:" + bid
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}

		defer file.Close()
		WriteFile(file, s)
	} else if err != nil {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		WriteFile(file, s)
		defer file.Close()
	}
}

func WriteFile(file *os.File, text string) {
	_, err := file.Write([]byte(text))
	if err != nil {
		panic(err)
	}
}
