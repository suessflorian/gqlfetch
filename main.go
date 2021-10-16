package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

//go:embed introspectQuery.graphql
var introspectSchema string

const DEFAULT_ENDPOINT = "http://localhost:8080/query"

func main() {
	var endpoint string
	headers := make(headers)

	flag.StringVar(&endpoint, "endpoint", DEFAULT_ENDPOINT, "GraphQL server endpoint")
	flag.Var(&headers, "header", "Headers to be passed endpoint (can appear multiple times)")
	flag.Parse()

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(graphQLRequest{Query: introspectSchema})
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest(http.MethodPost, endpoint, buffer)
	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header(headers)
	req.Header.Add("Content-Type", "application/json")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var schemaResponse introspectionResult
	err = json.NewDecoder(res.Body).Decode(&schemaResponse)
	if err != nil {
		log.Fatal(err)
	}

	if len(schemaResponse.Errors) != 0 {
		log.Fatal(schemaResponse.Errors)
    return
	}

	fmt.Println(printSchema(schemaResponse.Data.Schema))
}

type headers map[string][]string

func (h headers) Set(input string) error {
	values := strings.Split(input, "=")
	if len(values) < 2 {
		return errors.New(`header must appear like 'Authorization="Bearer token"'`)
	}
	h[values[0]] = append(h[values[0]], values[1:]...)
	return nil
}

func (h *headers) String() string {
	sb := strings.Builder{}
	for header, values := range *h {
		sb.WriteString(fmt.Sprintf("%s=%s", header, strings.Join(values, ",")))
	}
	return sb.String()
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlErrs []graphqlErr

type graphqlErr struct {
	Message string `json:"message"`
}
