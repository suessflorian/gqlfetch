package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/suessflorian/gqlfetch"
)


func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var filePath string
	var withoutBuiltins bool

	flag.BoolVar(&withoutBuiltins, "without-builtins", false, "Do not include builtin types")
	flag.StringVar(&filePath, "file", "schema.json", "Path to introspection file as json")
	flag.Parse()

	schema, err := gqlfetch.BuildClientSchemaFromFile(ctx, filePath, withoutBuiltins)
	if err != nil {
		panic(err)
	}
	fmt.Println(schema)
}