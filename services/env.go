package services

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port              string
	RestCountriesBase string
	CurrencyBase      string
}

// LoadConfig leser .env hvis den finnes, og henter verdier fra miljøet.
// Bruker sensible defaults for base-URLer hvis de ikke er satt.
// Ingen tredjepartsbibliotek.
func LoadConfig(dotenvPath string) (Config, error) {
	// 1) Last .env inn i prosessens env (om filen finnes)
	_ = loadDotEnv(dotenvPath)

	// 2) Hent verdier fra miljøvariabler, med defaults
	// Merk: Render setter PORT dynamisk; lokalt bruker vi 8080 hvis ikke satt.
	cfg := Config{
		Port: getenvDefault("PORT", "8080"),

		// Bruk kurs-URLene dere har fått i oppgaven som defaults.
		RestCountriesBase: strings.TrimRight(
			getenvDefault("RESTCOUNTRIES_BASE", "http://129.241.150.113:8080/v3.1"),
			"/",
		),
		CurrencyBase: strings.TrimRight(
			getenvDefault("CURRENCY_BASE", "http://129.241.150.113:9090/currency"),
			"/",
		),
	}

	// 3) Valider at base-URL-er finnes (de skal alltid finnes nå pga defaults)
	if cfg.RestCountriesBase == "" {
		return Config{}, fmt.Errorf("RESTCOUNTRIES_BASE mangler (sett i .env eller miljøvariabler)")
	}
	if cfg.CurrencyBase == "" {
		return Config{}, fmt.Errorf("CURRENCY_BASE mangler (sett i .env eller miljøvariabler)")
	}

	return cfg, nil
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// loadDotEnv leser enkel .env: KEY=VALUE, hopper over tomme linjer og #kommentarer.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		// Hvis .env ikke finnes, er det ok (vi bruker OS env / defaults).
		return nil
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Ikke overskriv hvis allerede satt i miljøet
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return sc.Err()
}
