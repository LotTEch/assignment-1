package main

import (
	"log"
	"net/http"
	"time"

	"assignment-1/handlers"
	"assignment-1/services"
)

func main() {
	// Les .env og/eller miljøvariabler
	cfg, err := services.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Kunne ikke laste config: %v", err)
	}

	// Brukes for uptime i /status
	startedAt := time.Now()

	// HTTP-klient med timeout for tredjepartskall
	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	// Services som snakker med tredjeparts API-er
	rc := services.NewRestCountriesService(cfg.RestCountriesBase, client)
	cc := services.NewCurrencyService(cfg.CurrencyBase, client)

	// Handlers (HTTP-laget)
	statusHandler := handlers.NewStatusHandler(startedAt, rc, cc)
	infoHandler := handlers.NewInfoHandler(rc)
	exchangeHandler := handlers.NewExchangeHandler(rc, cc)

	// Routes (endepunktene dine)
	mux := http.NewServeMux()
	mux.HandleFunc("/countryinfo/v1/status/", statusHandler.Handle)
	mux.HandleFunc("/countryinfo/v1/info/", infoHandler.Handle)
	mux.HandleFunc("/countryinfo/v1/exchange/", exchangeHandler.Handle)

	// Litt hjelpetekst på rot (valgfritt, men ok)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("CountryInfo service. Use /countryinfo/v1/status/, /info/{code}, /exchange/{code}\n"))
	})

	addr := ":" + cfg.Port
	log.Printf("Starter server på %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
