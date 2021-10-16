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

	fmt.Println(printSchema(schemaResponse.Data.Schema))
}

func printSchema(schema GraphQLSchema) string {
	sb := &strings.Builder{}

	printDirectives(sb, schema.Directives)
	sb.WriteString("\n")
	printTypes(sb, schema.Types)

	return sb.String()
}

func printDirectives(sb *strings.Builder, directives []Directives) {
	for _, directive := range directives {
		if directive.Description != "" {
			sb.WriteString(fmt.Sprintf(`"""%s"""`, directive.Description))
		}
		sb.WriteString(fmt.Sprintf("\ndirective @%s", directive.Name))
		if len(directive.Args) > 0 {
			sb.WriteString("(\n")
			for _, arg := range directive.Args {
				if arg.Description != "" {
					sb.WriteString(fmt.Sprintf(`"""%s"""`, arg.Description))
					sb.WriteString("\n")
				}
				sb.WriteString(fmt.Sprintf("%s: %s\n", arg.Name, printType(arg.Type)))
			}
			sb.WriteString("\n)")
		}

		sb.WriteString(" on ")
		for i, location := range directive.Locations {
			sb.WriteString(location)
			if i < len(directive.Locations)-1 {
				sb.WriteString(" | ")
			}
		}
	}
}

const (
	OBJECT       = "OBJECT"
	UNION        = "UNION"
	ENUM         = "ENUM"
	SCALAR       = "SCALAR"
	INPUT_OBJECT = "INPUT_OBJECT"
	INTERFACE    = "INTERFACE"
)

func printTypes(sb *strings.Builder, types []Types) {
	for _, typ := range types {
		if typ.Description != "" {
			sb.WriteString(fmt.Sprintf(`"""%s"""`, typ.Description))
			sb.WriteString("\n")
		}
		switch typ.Kind {
		case OBJECT:
			sb.WriteString(fmt.Sprintf("type %s {\n", typ.Name))
			for _, field := range typ.Fields {
				if typ.Description != "" {
					sb.WriteString(fmt.Sprintf(`"""%s"""`, typ.Description))
					sb.WriteString("\n")
				}
				sb.WriteString(fmt.Sprintf("%s: %s\n", field.Name, printType(field.Type)))
			}
			sb.WriteString("}")
		case UNION:
			sb.WriteString(fmt.Sprintf("union %s =", typ.Name))
			var possible []Type
			if err := json.Unmarshal(typ.PossibleTypes, &possible); err != nil {
				panic(err)
			}
			for i, typ := range possible {
				sb.WriteString(printType(typ))
				if i < len(possible)-1 {
					sb.WriteString(" | ")
				}
			}
		case ENUM:
			sb.WriteString(fmt.Sprintf("enum %s {\n", typ.Name))
			var enumValues []EnumValue
			if err := json.Unmarshal(typ.EnumValues, &enumValues); err != nil {
				panic(err)
			}
			for _, value := range enumValues {
				if typ.Description != "" {
					sb.WriteString(fmt.Sprintf(`"""%s"""`, typ.Description))
					sb.WriteString("\n")
				}
				sb.WriteString(fmt.Sprintf("%s\n", value.Name))
			}
			sb.WriteString("}")
		case SCALAR:
			sb.WriteString(fmt.Sprintf("scalar %s", typ.Name))
		case INPUT_OBJECT:
			sb.WriteString(fmt.Sprintf("input %s {\n", typ.Name))
			for _, field := range typ.Fields {
				if typ.Description != "" {
					sb.WriteString(fmt.Sprintf(`"""%s"""`, typ.Description))
					sb.WriteString("\n")
				}
				sb.WriteString(fmt.Sprintf("%s: %s\n", field.Name, printType(field.Type)))
			}
			sb.WriteString("}")
		case INTERFACE:
			sb.WriteString(fmt.Sprintf("interface %s {\n", typ.Name))
			for _, field := range typ.Fields {
				if typ.Description != "" {
					sb.WriteString(fmt.Sprintf(`"""%s"""`, typ.Description))
					sb.WriteString("\n")
				}
				sb.WriteString(fmt.Sprintf("%s: %s\n", field.Name, printType(field.Type)))
			}
			sb.WriteString("}")

		default:
			log.Println("not handling", typ.Kind)
		}
		sb.WriteString("\n")
	}
}

const (
	NON_NULL = "NON_NULL"
	LIST     = "LIST"
)

func printType(typ Type) string {
	var ofType Type
	if err := json.Unmarshal(typ.OfType, &ofType); err != nil {
		panic(err)
	}

	if ofType.isNil {
		return *typ.Name
	}

	switch typ.Kind {
	case NON_NULL:
		return fmt.Sprintf("%s!", printType(ofType))
	case LIST:
		return fmt.Sprintf("[%s]", printType(ofType))
	default:
		panic(fmt.Sprintf("do not recognize type kind: %q", typ.Kind))
	}
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
		Schema GraphQLSchema `json:"__schema"`
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

type GraphQLSchema struct {
	QueryType struct {
		Name string `json:"name"`
	} `json:"queryType"`
	MutationType struct {
		Name string `json:"name"`
	} `json:"mutationType"`
	SubscriptionType interface{}  `json:"subscriptionType"`
	Types            []Types      `json:"types"`
	Directives       []Directives `json:"directives"`
}

type EnumValue struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Types struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Fields      []struct {
		Name              string        `json:"name"`
		Description       string        `json:"description"`
		Args              []interface{} `json:"args"`
		Type              Type          `json:"type"`
		IsDeprecated      bool          `json:"isDeprecated"`
		DeprecationReason interface{}   `json:"deprecationReason"`
	} `json:"fields"`
	InputFields   []InputField    `json:"inputFields"`
	Interfaces    json.RawMessage `json:"interfaces"`
	EnumValues    json.RawMessage `json:"enumValues"`
	PossibleTypes json.RawMessage `json:"possibleTypes"`
}

type Interface struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type InputField struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         Type        `json:"type"`
	DefaultValue interface{} `json:"defaultValue"`
}

type Directives struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Locations   []string `json:"locations"`
	Args        []struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Type         `json:"type"`
		DefaultValue interface{} `json:"defaultValue"`
	} `json:"args"`
}

type Type struct {
	typ
	isNil bool
}

type typ struct {
	Kind   string          `json:"kind"`
	Name   *string         `json:"name"`
	OfType json.RawMessage `json:"ofType"`
}

func (t *Type) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.isNil = true
	}

	var typ typ
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}

	t.Kind = typ.Kind
	t.Name = typ.Name
	t.OfType = typ.OfType

	return nil
}
