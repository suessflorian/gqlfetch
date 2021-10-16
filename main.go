package gqlfetch

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed introspectQuery.graphql
var introspectSchema string
type ServerConfig struct {
	Endpoint string
	Headers  http.Header
}

func BuildClientSchema(ctx context.Context, cfg ServerConfig) (string, error) {
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(struct{ Query string }{Query: introspectSchema}); err != nil {
		return "", fmt.Errorf("failed to prepare introspection query request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Endpoint, buffer)
	if err != nil {
		return "", fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header = http.Header(cfg.Headers)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{Timeout: 2 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var schemaResponse introspectionResults
	err = json.NewDecoder(res.Body).Decode(&schemaResponse)
	if err != nil {
		log.Fatal(err)
	}

	if len(schemaResponse.Errors) != 0 {
		var errs []string
		for _, err := range schemaResponse.Errors {
			errs = append(errs, err.Message)
		}
		return "", errors.New("encountered the following GraphQL errors: " + strings.Join(errs, ","))
	}

	return printSchema(schemaResponse.Data.Schema), nil
}
