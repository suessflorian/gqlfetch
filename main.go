package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed introspectQuery.graphql
var introspectSchema string

func main() {
	ctx := context.Background()
	endpoint := os.Getenv("SERVER_ENDPOINT")
	if strings.TrimSpace(endpoint) == "" {
		log.Fatal("SERVER_ENDPOINT must be provided")
	}

	authorization := os.Getenv("AUTHORIZATION_HEADER")

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(graphQLRequest{Query: introspectSchema})
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, buffer)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", authorization)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{Timeout: 2 * time.Minute}
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
	}

	fmt.Println(printSchema(schemaResponse.Data.Schema))
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlErrs []graphqlErr

type graphqlErr struct {
	Message string `json:"message"`
}
