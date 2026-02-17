package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"assignment-1/services"
)

type StatusHandler struct {
	startedAt time.Time
	rc        *services.RestCountriesService
	cc        *services.CurrencyService
}

func NewStatusHandler(startedAt time.Time, rc *services.RestCountriesService, cc *services.CurrencyService) *StatusHandler {
	return &StatusHandler{startedAt: startedAt, rc: rc, cc: cc}
}

func (h *StatusHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Sjekk RestCountries ved å kalle et kjent, lett endepunkt (no).
	_, rcStatus, rcErr := h.rc.GetCountryByCode("no")
	if rcErr != nil && rcStatus == 0 {
		// hvis nettverksfeil før statuskode, velg 503
		rcStatus = http.StatusServiceUnavailable
	}

	// Sjekk Currency API ved å kalle NOK (lett og stabilt).
	_, ccStatus, ccErr := h.cc.GetRates("NOK")
	if ccErr != nil && ccStatus == 0 {
		ccStatus = http.StatusServiceUnavailable
	}

	resp := map[string]any{
		"restcountriesapi": rcStatus,
		"currenciesapi":    ccStatus,
		"version":          "v1",
		"uptime":           int(time.Since(h.startedAt).Seconds()),
	}

	// Hvis alt OK → 200, ellers 503 (enkelt og greit)
	status := http.StatusOK
	if rcStatus != http.StatusOK || ccStatus != http.StatusOK {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
