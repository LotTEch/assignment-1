package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RestCountriesService struct {
	baseURL string
	client  *http.Client
}

func NewRestCountriesService(baseURL string, client *http.Client) *RestCountriesService {
	return &RestCountriesService{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

// Country representerer bare feltene vi trenger (ikke hele monster-responsen).
type Country struct {
	Name struct {
		Common string `json:"common"`
	} `json:"name"`

	Continents []string          `json:"continents"`
	Population int               `json:"population"`
	Area       float64           `json:"area"`
	Languages  map[string]string `json:"languages"`
	Borders    []string          `json:"borders"`
	Capital    []string          `json:"capital"`
	Flags      struct {
		PNG string `json:"png"`
	} `json:"flags"`

	// currencies er et map der nøklene er ISO4217-koder (NOK, EUR, SEK...)
	Currencies map[string]struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"currencies"`
}

func (s *RestCountriesService) GetCountryByCode(code string) (Country, int, error) {
	url := fmt.Sprintf("%s/alpha/%s", s.baseURL, strings.ToLower(code))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Country{}, 0, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return Country{}, 0, err
	}
	defer resp.Body.Close()

	// RestCountries returnerer ofte en liste med 1 land
	if resp.StatusCode != http.StatusOK {
		return Country{}, resp.StatusCode, fmt.Errorf("restcountries status %d", resp.StatusCode)
	}

	var arr []Country
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return Country{}, resp.StatusCode, err
	}
	if len(arr) == 0 {
		return Country{}, http.StatusNotFound, fmt.Errorf("ingen land funnet")
	}

	return arr[0], resp.StatusCode, nil
}

// Helper: trekk ut "første" valutakode (NOK for Norge).
func FirstCurrencyCode(c Country) (string, bool) {
	for code := range c.Currencies {
		return code, true
	}
	return "", false
}
