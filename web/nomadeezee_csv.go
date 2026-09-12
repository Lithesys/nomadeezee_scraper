package web

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var nomadeezeePlaceCSVHeaders = []string{
	"title",
	"description",
	"category",
	"latitude",
	"longitude",
	"address",
	"country",
	"province",
	"is_public",
	"is_premium_only",
}

type scrapedCompleteAddress struct {
	City    string `json:"city"`
	Country string `json:"country"`
	State   string `json:"state"`
}

// writeNomadeezeeCSV adapts the scraper's CSV schema to Nomadeezee's
// public.places insert shape. Database-owned fields (id, user_id, timestamps)
// are intentionally omitted so Supabase assigns them safely.
func writeNomadeezeeCSV(dst io.Writer, src io.Reader) error {
	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read source header: %w", err)
	}

	columns := make(map[string]int, len(header))
	for i, name := range header {
		columns[name] = i
	}

	value := func(row []string, name string) string {
		index, ok := columns[name]
		if !ok || index >= len(row) {
			return ""
		}

		return row[index]
	}

	writer := csv.NewWriter(dst)
	if err := writer.Write(nomadeezeePlaceCSVHeaders); err != nil {
		return fmt.Errorf("write output header: %w", err)
	}

	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read source row: %w", readErr)
		}

		country, province := nomadeezeeLocation(value(row, "complete_address"))
		if country == "" {
			country = "Vietnam"
		}

		description := value(row, "descriptions")
		if description == "" {
			description = value(row, "description")
		}

		if err := writer.Write([]string{
			value(row, "title"),
			description,
			nomadeezeeCategory(value(row, "category")),
			value(row, "latitude"),
			value(row, "longitude"),
			value(row, "address"),
			country,
			province,
			"true",
			"false",
		}); err != nil {
			return fmt.Errorf("write output row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}

	return nil
}

func nomadeezeeLocation(value string) (string, string) {
	var address scrapedCompleteAddress
	if err := json.Unmarshal([]byte(value), &address); err != nil {
		return "", ""
	}

	province := address.State
	if province == "" {
		province = address.City
	}

	return address.Country, province
}

func nomadeezeeCategory(value string) string {
	category := strings.ToLower(value)

	switch {
	case strings.Contains(category, "cafe"), strings.Contains(category, "coffee"):
		return "cafe"
	case strings.Contains(category, "restaurant"), strings.Contains(category, "food"), strings.Contains(category, "bakery"):
		return "cuisine"
	case strings.Contains(category, "hotel"), strings.Contains(category, "hostel"), strings.Contains(category, "resort"):
		return "accommodation"
	case strings.Contains(category, "museum"), strings.Contains(category, "gallery"):
		return "museum"
	case strings.Contains(category, "beach"):
		return "beach"
	case strings.Contains(category, "bar"), strings.Contains(category, "pub"), strings.Contains(category, "nightclub"):
		return "nightlife"
	case strings.Contains(category, "temple"), strings.Contains(category, "church"), strings.Contains(category, "mosque"):
		return "religious"
	case strings.Contains(category, "spa"), strings.Contains(category, "wellness"), strings.Contains(category, "gym"):
		return "wellness"
	case strings.Contains(category, "park"), strings.Contains(category, "nature"), strings.Contains(category, "waterfall"), strings.Contains(category, "mountain"):
		return "nature"
	case strings.Contains(category, "shop"), strings.Contains(category, "store"), strings.Contains(category, "market"), strings.Contains(category, "mall"):
		return "shopping"
	case strings.Contains(category, "cinema"), strings.Contains(category, "theatre"), strings.Contains(category, "amusement"):
		return "entertainment"
	case strings.Contains(category, "tour"), strings.Contains(category, "adventure"):
		return "adventure"
	case strings.Contains(category, "histor"), strings.Contains(category, "heritage"):
		return "historical"
	case strings.Contains(category, "sightseeing"), strings.Contains(category, "landmark"):
		return "sightseeing"
	default:
		return "other"
	}
}
