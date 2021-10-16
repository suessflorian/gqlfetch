package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/suessflorian/gqlfetch"
)

const DEFAULT_ENDPOINT = "http://localhost:8080/query"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var endpoint string
	headers := make(headers)

	flag.StringVar(&endpoint, "endpoint", DEFAULT_ENDPOINT, "GraphQL server endpoint")
	flag.Var(&headers, "header", "Headers to be passed endpoint (can appear multiple times)")
	flag.Parse()

	fmt.Println(gqlfetch.BuildClientSchemaWithHeaders(ctx, endpoint, http.Header(headers)))
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
