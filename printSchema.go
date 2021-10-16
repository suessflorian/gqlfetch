package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/ast"
)

func printSchema(schema introspectionSchema) string {
	sb := &strings.Builder{}

	printDirectives(sb, schema.Directives)
	printTypes(sb, schema.Types)

	return sb.String()
}

func printDirectives(sb *strings.Builder, directives []introspectionDirectiveDefinition) {
	for _, directive := range directives {
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

func printTypes(sb *strings.Builder, types []introspectionTypeDefinition) {
	for _, typ := range types {
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
				sb.WriteString(fmt.Sprintf("\t%s: %s\n", field.Name, introspectionTypeToAstType(field.Type).String()))
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
			for _, field := range typ.Fields {
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
