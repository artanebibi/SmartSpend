package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/otiai10/gosseract/v2"
)

func OcrHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, file)

	client := gosseract.NewClient()
	defer client.Close()

	client.SetLanguage("mkd")
	client.SetPageSegMode(4)

	client.SetImageFromBytes(buf.Bytes())

	data, err := client.Text()
	if err != nil {
		http.Error(w, fmt.Sprintf("OCR error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"data": %q}`, data)))
}
