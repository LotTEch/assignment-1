package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type CurrencyService struct {
	baseURL string
	client  *http.Client
}

func NewCurrencyService(baseURL string, client *http.Client) *CurrencyService {
	return &CurrencyService{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

type CurrencyResponse struct {
	Result   string             `json:"result"`
	BaseCode string             `json:"base_code"`
	Rates    map[string]float64 `json:"rates"`
}

func (s *CurrencyService) GetRates(baseCurrency string) (CurrencyResponse, int, error) {
	url := fmt.Sprintf("%s/%s", s.baseURL, strings.ToUpper(baseCurrency))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return CurrencyResponse{}, 0, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return CurrencyResponse{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CurrencyResponse{}, resp.StatusCode, fmt.Errorf("currency api status %d", resp.StatusCode)
	}

	var cr CurrencyResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return CurrencyResponse{}, resp.StatusCode, err
	}

	// API-et kan returnere JSON med result=error
	if strings.ToLower(cr.Result) != "success" {
		return CurrencyResponse{}, resp.StatusCode, fmt.Errorf("currency api result=%s", cr.Result)
	}

	return cr, resp.StatusCode, nil
}
