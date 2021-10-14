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

//go:embed introspectionQuery.graphql
var introspectionQuery string

func main() {
	ctx := context.Background()
	endpoint := os.Getenv("SERVER_ENDPOINT")
	if strings.TrimSpace(endpoint) == "" {
		log.Fatal("SERVER_ENDPOINT must be provided")
	}

	authorization := os.Getenv("AUTHORIZATION_HEADER")

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(graphQLRequest{Query: introspectionQuery})
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

	var schemaResponse introspectionRes
	err = json.NewDecoder(res.Body).Decode(&schemaResponse)
	if err != nil {
		log.Fatal(err)
	}

	if len(schemaResponse.Errors) != 0 {
		log.Fatal(schemaResponse.Errors)
	}

	fmt.Println(schemaResponse)
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlErrs []graphqlErr

type graphqlErr struct {
	Message string `json:"message"`
}

type introspectionRes struct {
	Errors graphqlErrs `json:"errors"`
	Data   struct {
		Schema struct {
			QueryType struct {
				Name string `json:"name"`
			} `json:"queryType"`
			MutationType struct {
				Name string `json:"name"`
			} `json:"mutationType"`
			SubscriptionType interface{} `json:"subscriptionType"`
			Types            []struct {
				Kind        string `json:"kind"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Fields      []struct {
					Name        string        `json:"name"`
					Description string        `json:"description"`
					Args        []interface{} `json:"args"`
					Type        struct {
						Kind   string      `json:"kind"`
						Name   interface{} `json:"name"`
						OfType struct {
							Kind   string      `json:"kind"`
							Name   string      `json:"name"`
							OfType interface{} `json:"ofType"`
						} `json:"ofType"`
					} `json:"type"`
					IsDeprecated      bool        `json:"isDeprecated"`
					DeprecationReason interface{} `json:"deprecationReason"`
				} `json:"fields"`
				InputFields []struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					Type        struct {
						Kind   string      `json:"kind"`
						Name   interface{} `json:"name"`
						OfType struct {
							Kind   string      `json:"kind"`
							Name   string      `json:"name"`
							OfType interface{} `json:"ofType"`
						} `json:"ofType"`
					} `json:"type"`
					DefaultValue interface{} `json:"defaultValue"`
				} `json:"inputFields"`
				Interfaces    []interface{} `json:"interfaces"`
				EnumValues    []interface{} `json:"enumValues"`
				PossibleTypes interface{}   `json:"possibleTypes"`
			} `json:"types"`
			Directives []struct {
				Name        string   `json:"name"`
				Description string   `json:"description"`
				Locations   []string `json:"locations"`
				Args        []struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					Type        struct {
						Kind   string      `json:"kind"`
						Name   interface{} `json:"name"`
						OfType struct {
							Kind   string      `json:"kind"`
							Name   string      `json:"name"`
							OfType interface{} `json:"ofType"`
						} `json:"ofType"`
					} `json:"type"`
					DefaultValue interface{} `json:"defaultValue"`
				} `json:"args"`
			} `json:"directives"`
		} `json:"__schema"`
	} `json:"data"`
	Extensions struct {
		Plan struct {
			RootSteps interface{} `json:"RootSteps"`
		} `json:"plan"`
		Timing struct {
			Execution string `json:"execution"`
			Format    string `json:"format"`
			Merge     string `json:"merge"`
		} `json:"timing"`
	} `json:"extensions"`
}
