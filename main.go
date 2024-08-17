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

//go:embed introspect.graphql
var introspectSchema string

type BuildClientSchemaOptions struct {
	Endpoint        string
	Method          string
	Headers         http.Header
	WithoutBuiltins bool
}

func BuildClientSchema(ctx context.Context, endpoint string, withoutBuiltins bool) (string, error) {
	return BuildClientSchemaWithOptions(ctx, BuildClientSchemaOptions{
		Endpoint:        endpoint,
		Method:          http.MethodPost,
		Headers:         make(http.Header),
		WithoutBuiltins: withoutBuiltins,
	})
}

func BuildClientSchemaWithHeaders(ctx context.Context, endpoint string, headers http.Header, withoutBuiltins bool) (string, error) {
	return BuildClientSchemaWithOptions(ctx, BuildClientSchemaOptions{
		Endpoint:        endpoint,
		Method:          http.MethodPost,
		Headers:         headers,
		WithoutBuiltins: withoutBuiltins,
	})
}

func BuildClientSchemaWithOptions(ctx context.Context, options BuildClientSchemaOptions) (string, error) {
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(struct {
		Query string `json:"query"`
	}{Query: introspectSchema}); err != nil {
		return "", fmt.Errorf("failed to prepare introspection query request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, options.Method, options.Endpoint, buffer)
	if err != nil {
		return "", fmt.Errorf("failed to create query request: %w", err)
	}

	// If no headers are provided, create an empty header map, so we can add the content type header
	if options.Headers == nil {
		options.Headers = make(http.Header)
	}
	req.Header = http.Header(options.Headers)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{Timeout: 2 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to download schema: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unable to download schema: %s", res.Status)
	}

	var schemaResponse introspectionResults
	err = json.NewDecoder(res.Body).Decode(&schemaResponse)
	if err != nil {
		return "", fmt.Errorf("unable to decode schema: %w", err)
	}

	if len(schemaResponse.Errors) != 0 {
		var errs []string
		for _, err := range schemaResponse.Errors {
			errs = append(errs, err.Message)
		}
		return "", errors.New("encountered the following GraphQL errors: " + strings.Join(errs, ","))
	}

	schema := schemaResponse.Data.Schema

	normalizeSchema(&schema)

	return printSchema(schema, options.WithoutBuiltins), nil
}

func normalizeSchema(schema *introspectionSchema) {
	schema.Directives = uniqDirectives(schema.Directives)
}

func uniqDirectives(directives []introspectionDirectiveDefinition) []introspectionDirectiveDefinition {
	nameToDrctv := map[string][]introspectionDirectiveDefinition{}
	for _, d := range directives {
		nameToDrctv[d.Name] = append(nameToDrctv[d.Name], d)
	}

	for name, directives := range nameToDrctv {
		uniq := []introspectionDirectiveDefinition{}
		for _, d := range directives {
			duplicateFound := false
			for _, ud := range uniq {
				if d.Equal(ud) {
					duplicateFound = true
					break
				}
			}
			if !duplicateFound {
				uniq = append(uniq, d)
			}
		}

		if len(uniq) > 1 {
			log.Printf("directives have same name but different propertes: %s", name)
		}

		nameToDrctv[name] = uniq
	}

	res := []introspectionDirectiveDefinition{}
	for _, ds := range nameToDrctv {
		res = append(res, ds...)
	}
	return res
}
