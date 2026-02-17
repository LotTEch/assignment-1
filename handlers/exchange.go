package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"assignment-1/services"
)

type ExchangeHandler struct {
	rc *services.RestCountriesService
	cc *services.CurrencyService
}

func NewExchangeHandler(rc *services.RestCountriesService, cc *services.CurrencyService) *ExchangeHandler {
	return &ExchangeHandler{rc: rc, cc: cc}
}

func (h *ExchangeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code, ok := parseExchangeCountryCode(r.URL.Path)
	if !ok {
		http.Error(w, "two_letter_country_code må være 2 bokstaver", http.StatusBadRequest)
		return
	}

	baseCountry, baseCurrency, errStatus, err := h.loadBaseCountry(code)
	if err != nil {
		http.Error(w, "feil ved henting av base-land", errStatus)
		return
	}

	exchangeRates, errStatus, err := h.buildExchangeRates(baseCountry, baseCurrency)
	if err != nil {
		http.Error(w, "feil ved henting av valutakurser", errStatus)
		return
	}

	out := map[string]any{
		"country":        baseCountry.Name.Common,
		"base-currency":  baseCurrency,
		"exchange-rates": exchangeRates,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

// parseExchangeCountryCode henter ut {code} fra pathen /countryinfo/v1/exchange/{code}
// og validerer at den er 2 bokstaver.
func parseExchangeCountryCode(path string) (string, bool) {
	prefix := "/countryinfo/v1/exchange/"
	code := strings.TrimPrefix(path, prefix)
	code = strings.Trim(code, "/")

	if len(code) != 2 {
		return "", false
	}
	return code, true
}

// loadBaseCountry henter base-landet fra RestCountries og finner basevaluta (ISO 4217)
// basert på nøkkelen i "currencies"-objektet.
func (h *ExchangeHandler) loadBaseCountry(code string) (services.Country, string, int, error) {
	baseCountry, status, err := h.rc.GetCountryByCode(code)
	if err != nil {
		if status == 0 {
			status = http.StatusServiceUnavailable
		}
		return services.Country{}, "", status, err
	}

	baseCurrency, ok := services.FirstCurrencyCode(baseCountry)
	if !ok {
		return services.Country{}, "", http.StatusBadGateway, fmt.Errorf("ingen base currency funnet")
	}

	return baseCountry, baseCurrency, http.StatusOK, nil
}

// buildExchangeRates bygger listen "exchange-rates" ved å:
// 1) Hente alle valutakurser for baseCurrency fra Currency API én gang
// 2) Slå opp hvert naboland via RestCountries (alpha-kode fra borders)
// 3) Finne nabolandets valuta-kode (første currency key)
// 4) Plukke riktig kurs fra rates-tabellen og bygge [{ "EUR": 0.08 }, ...]
func (h *ExchangeHandler) buildExchangeRates(baseCountry services.Country, baseCurrency string) ([]map[string]float64, int, error) {
	ratesResp, status, err := h.cc.GetRates(baseCurrency)
	if err != nil {
		if status == 0 {
			status = http.StatusServiceUnavailable
		}
		return nil, status, err
	}

	exchangeRates := make([]map[string]float64, 0)

	for _, borderCode := range baseCountry.Borders {
		neighbor, nStatus, nErr := h.rc.GetCountryByCode(borderCode)
		if nErr != nil {
			if nStatus == 0 {
				nStatus = http.StatusBadGateway
			}
			return nil, nStatus, nErr
		}

		neighborCurrency, ok := services.FirstCurrencyCode(neighbor)
		if !ok {
			// Hvis naboland mangler valuta, hopp over
			continue
		}

		rate, exists := ratesResp.Rates[neighborCurrency]
		if !exists {
			// Hvis Currency API ikke har denne valutaen, hopp over
			continue
		}

		exchangeRates = append(exchangeRates, map[string]float64{
			neighborCurrency: rate,
		})
	}

	return exchangeRates, http.StatusOK, nil
}
