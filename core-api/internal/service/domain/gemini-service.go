package domain

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/domain/model"
	"SmartSpend/internal/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	database_          database.Service               = database.New()
	categoryRepository repository.ICategoryRepository = repository.NewCategoryRepository(database_)
	categoryService    ICategoryService               = NewCategoryService(categoryRepository)
	prompt                                            = fmt.Sprintf("Based on the image, which is a receipt, try to create me a Transaction json.\nRules:\n\nI suggest you create an array of items (item as a key, and price as a value so you have it easier to calculate the total after) but do not include it in the response.\nHow the response should look like:\n{\n\"id\": 0, // leave 0, its autoincremented\n\"title\": \"\", //Based on the items which you have extracted suggest me a title for the transaction made in English\n\"price\": \"\", //Calculate the total price from the receipt\n\"date_made\": \"0001-01-01T11:11:05Z\",\n\"owner_id\": \"\",\n\"category_id\": \"From the list of categories: %v choose one where you think the current transaction falls best into, but add the id\"\n\"type: \"Expense\"\n}\nDo not let your model fail to prioritize a semantically correct and common product name over a literal, but flawed, character transcription\n", categoryService.FindAll())
	promptOCR                                         = fmt.Sprintf("Based on the data extracted below from an OCR service in Tesseract in Macedonian try to extract the item names, if multiple items are tried to be written but in different matter (letters are shuffled) try to predict / find the real item in Macedonian Markets.\nRules:\n1. Output in JSON only.\n2. Create me a Transaction model which JSON looks like this\n3. I suggest you create an array of items (item as a key, and price as a value so you have it easier to calculate the total after) but do not include it in the response.\nHow the response should look like:\n{\n\"id\": 0, // leave 0, its autoincremented\n\"title\": \"\", //Based on the items which you have extracted suggest me a title for the transaction made in English\n\"price\": \"\", //Calculate the total price from the receipt\n\"date_made\": \"\",\n\"owner_id\": \"\",\n\"category_id\": \"From the list of categories: %v choose one where you think the current transaction falls best into, but add the id\"\n\"type: \"Expense\"\n}\n4. Only include items that make sense in a Macedonian market.\n5. Dont just trust the text blindly, if there are multiple of the 'same' items display them in the result.\n6. If there are '{number}x' before of what you think is an Item, multiply the price and update the quantity accordingly.\nDo not let your model fail to prioritize a semantically correct and common product name over a literal, but flawed, character transcription\n", categoryService.FindAll())
)

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type IGeminiService interface {
	SendToGemini(extractedTextOCR string, imageString string) (*model.Transaction, error)
}

type GeminiService struct {
	apiKey string
}

func NewGeminiService() *GeminiService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	return &GeminiService{apiKey: apiKey}
}

func (g *GeminiService) SendToGemini(extractedTextOCR string, imageString string) (*model.Transaction, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
					{
						"inline_data": map[string]string{
							"mime_type": "image/jpeg",
							"data":      imageString,
						},
					},
				},
			},
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s?key=%s", url, g.apiKey), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gemini API response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini API response did not contain a valid candidate with text content")
	}

	jsonText := geminiResp.Candidates[0].Content.Parts[0].Text
	jsonText = strings.TrimSpace(jsonText)
	jsonText = strings.TrimPrefix(jsonText, "```json\n")
	jsonText = strings.TrimSuffix(jsonText, "\n```")

	var tx model.Transaction
	if err := json.Unmarshal([]byte(jsonText), &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extracted JSON text to transaction model: %w", err)
	}

	return &tx, nil
}
