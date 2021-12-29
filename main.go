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

	"github.com/vektah/gqlparser/ast"
)

//go:embed introspect.graphql
var introspectSchema string

func BuildClientSchema(ctx context.Context, endpoint string, withoutBuiltins bool) (string, error) {
	return BuildClientSchemaWithHeaders(ctx, endpoint, make(http.Header), withoutBuiltins)
}

func BuildClientSchemaWithHeaders(ctx context.Context, endpoint string, headers http.Header, withoutBuiltins bool) (string, error) {
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(struct {
		Query string `json:"query"`
	}{Query: introspectSchema}); err != nil {
		return "", fmt.Errorf("failed to prepare introspection query request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, buffer)
	if err != nil {
		return "", fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header = http.Header(headers)
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

	return printSchema(schemaResponse.Data.Schema, withoutBuiltins), nil
}

func printSchema(schema introspectionSchema, withoutBuiltins bool) string {
	sb := &strings.Builder{}

	printDirectives(sb, schema.Directives, withoutBuiltins)
	printTypes(sb, schema.Types, withoutBuiltins)

	return sb.String()
}

func printDirectives(sb *strings.Builder, directives []introspectionDirectiveDefinition, withoutBuiltins bool) {
	for _, directive := range directives {
		if withoutBuiltins && containsStr(directive.Name, excludeDirectives) {
			continue
		}
		printDescription(sb, directive.Description)
		sb.WriteString(fmt.Sprintf("directive @%s", directive.Name))
		if len(directive.Args) > 0 {
			sb.WriteString("(\n")
			for _, arg := range directive.Args {
				printDescription(sb, arg.Description)
				sb.WriteString(fmt.Sprintf("\t%s: %s\n", arg.Name, introspectionTypeToAstType(arg.Type).String()))
			}
			sb.WriteString(")")
		}

		sb.WriteString(" on ")
		for i, location := range directive.Locations {
			sb.WriteString(string(location))
			if i < len(directive.Locations)-1 {
				sb.WriteString(" | ")
			}
		}
		sb.WriteString("\n")
		sb.WriteString("\n")
	}
}

func printTypes(sb *strings.Builder, types []introspectionTypeDefinition, withoutBuiltins bool) {
	for _, typ := range types {
		if strings.HasPrefix(typ.Name, "__") {
			continue
		}
		if withoutBuiltins && containsStr(typ.Name, excludeScalarTypes) && typ.Kind == ast.Scalar {
			continue
		}
		printDescription(sb, typ.Description)

		switch typ.Kind {

		case ast.Object:
			sb.WriteString(fmt.Sprintf("type %s ", typ.Name))
			if len(typ.Interfaces) > 0 {
				sb.WriteString("implements ")
				for i, intface := range typ.Interfaces {
					sb.WriteString(intface.Name)
					if i < len(typ.Interfaces)-1 {
						sb.WriteString(" & ")
					}
				}
			}
			sb.WriteString("{\n")
			for _, field := range typ.Fields {
				printDescription(sb, field.Description)
				sb.WriteString(fmt.Sprintf("\t%s", field.Name))
				if len(field.Args) > 0 {
					sb.WriteString("(\n")
					for _, arg := range field.Args {
						printDescription(sb, arg.Description)
						sb.WriteString(fmt.Sprintf("\t\t%s: %s\n", arg.Name, introspectionTypeToAstType(arg.Type).String()))
					}
					sb.WriteString("\t)")
				}
				sb.WriteString(fmt.Sprintf(": %s\n", introspectionTypeToAstType(field.Type).String()))
			}
			sb.WriteString("}")

		case ast.Union:
			sb.WriteString(fmt.Sprintf("union %s =", typ.Name))
			var possible []*introspectedType
			if err := json.Unmarshal(typ.PossibleTypes, &possible); err != nil {
				panic(err)
			}
			for i, typ := range possible {
				sb.WriteString(introspectionTypeToAstType(typ).String())
				if i < len(possible)-1 {
					sb.WriteString(" | ")
				}
			}

		case ast.Enum:
			sb.WriteString(fmt.Sprintf("enum %s {\n", typ.Name))
			var enumValues ast.EnumValueList
			if err := json.Unmarshal(typ.EnumValues, &enumValues); err != nil {
				panic(err)
			}
			for _, value := range enumValues {
				printDescription(sb, value.Description)
				sb.WriteString(fmt.Sprintf("\t%s\n", value.Name))
			}
			sb.WriteString("}")

		case ast.Scalar:
			sb.WriteString(fmt.Sprintf("scalar %s", typ.Name))

		case ast.InputObject:
			sb.WriteString(fmt.Sprintf("input %s {\n", typ.Name))
			for _, field := range typ.InputFields {
				printDescription(sb, typ.Description)
				sb.WriteString(fmt.Sprintf("\t%s: %s\n", field.Name, introspectionTypeToAstType(field.Type).String()))
			}
			sb.WriteString("}")

		case ast.Interface:
			sb.WriteString(fmt.Sprintf("interface %s {\n", typ.Name))
			for _, field := range typ.Fields {
				printDescription(sb, typ.Description)
				sb.WriteString(fmt.Sprintf("\t%s: %s\n", field.Name, introspectionTypeToAstType(field.Type).String()))
			}
			sb.WriteString("}")

		default:
			panic(fmt.Sprint("not handling", typ.Kind))
		}
		sb.WriteString("\n")
		sb.WriteString("\n")
	}
}

func printDescription(sb *strings.Builder, description string) {
	if description != "" {
		sb.WriteString(fmt.Sprintf(`"""%s"""`, description))
		sb.WriteString("\n")
	}
}
