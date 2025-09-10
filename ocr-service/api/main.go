package main

import (
	"log"
	"net/http"
	"testOCR/handlers"
)

func main() {
	http.HandleFunc("/ocr", handlers.OcrHandler) // not using Gin, need it to be as light as possible!
	log.Fatal(http.ListenAndServe(":5000", nil))
}
