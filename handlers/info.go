package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"assignment-1/services"
)

type InfoHandler struct {
	rc *services.RestCountriesService
}

func NewInfoHandler(rc *services.RestCountriesService) *InfoHandler {
	return &InfoHandler{rc: rc}
}

func (h *InfoHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Path er /countryinfo/v1/info/{code}
	prefix := "/countryinfo/v1/info/"
	code := strings.TrimPrefix(r.URL.Path, prefix)
	code = strings.Trim(code, "/")

	if len(code) != 2 {
		http.Error(w, "two_letter_country_code må være 2 bokstaver", http.StatusBadRequest)
		return
	}

	c, status, err := h.rc.GetCountryByCode(code)
	if err != nil {
		// Propager “riktig” feil så godt vi kan
		if status == 0 {
			status = http.StatusServiceUnavailable
		}
		http.Error(w, "feil ved henting av landinfo", status)
		return
	}

	capital := ""
	if len(c.Capital) > 0 {
		capital = c.Capital[0]
	}

	out := map[string]any{
		"name":       c.Name.Common,
		"continents": c.Continents,
		"population": c.Population,
		"area":       c.Area,
		"languages":  c.Languages,
		"borders":    c.Borders,
		"flag":       c.Flags.PNG,
		"capital":    capital,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}
