package gqlfetch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

func printSchema(schema introspectionSchema, withoutBuiltins bool) string {
	sb := &strings.Builder{}

	err := printDirectives(sb, schema.Directives, withoutBuiltins)
	if err != nil {
		return fmt.Sprintf("unable to write directives: %v", err)
	}
	err = printTypes(sb, schema.Types, withoutBuiltins)
	if err != nil {
		return fmt.Sprintf("unable to write types: %v", err)
	}

	return sb.String()
}

func printDirectives(sb *strings.Builder, directives []introspectionDirectiveDefinition, withoutBuiltins bool) error {
	for _, directive := range directives {
		if withoutBuiltins && containsStr(directive.Name, excludeDirectives) {
			continue
		}
		err := printDescription(sb, directive.Description)
		if err != nil {
			return fmt.Errorf("unable to write directive description for %s: %w", directive.Name, err)
		}
		sb.WriteString(fmt.Sprintf("directive @%s", directive.Name))
		if len(directive.Args) > 0 {
			sb.WriteString("(\n")
			for _, arg := range directive.Args {
				err = printDescription(sb, arg.Description)
				if err != nil {
					return fmt.Errorf("unable to write description for arg %s.%s: %w", directive.Name, arg.Name, err)
				}
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
	return nil
}

func printTypes(sb *strings.Builder, types []introspectionTypeDefinition, withoutBuiltins bool) error {
	for _, typ := range types {
		if strings.HasPrefix(typ.Name, "__") {
			continue
		}
		if withoutBuiltins && containsStr(typ.Name, excludeScalarTypes) && typ.Kind == ast.Scalar {
			continue
		}
		err := printDescription(sb, typ.Description)
		if err != nil {
			return fmt.Errorf("unable to write description for type %s: %w", typ.Name, err)
		}

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
				err = printDescription(sb, field.Description)
				if err != nil {
					return fmt.Errorf("unable to write description for field %s.%s: %w", typ.Name, field.Name, err)
				}
				sb.WriteString(fmt.Sprintf("\t%s", field.Name))
				if len(field.Args) > 0 {
					sb.WriteString("(\n")
					for _, arg := range field.Args {
						err = printDescription(sb, arg.Description)
						if err != nil {
							return fmt.Errorf("unable to write description for arg %s.%s.%s: %w", typ.Name, field.Name, arg.Name, err)
						}
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
				return fmt.Errorf("unable to unmarshal possible types for %s: %w", typ.Name, err)
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
				return fmt.Errorf("unable to unmarshal enum values for %s: %w", typ.Name, err)
			}
			for _, value := range enumValues {
				err = printDescription(sb, value.Description)
				if err != nil {
					return fmt.Errorf("unable to write description for enum value %s.%s: %w", typ.Name, value.Name, err)
				}
				sb.WriteString(fmt.Sprintf("\t%s\n", value.Name))
			}
			sb.WriteString("}")

		case ast.Scalar:
			sb.WriteString(fmt.Sprintf("scalar %s", typ.Name))

		case ast.InputObject:
			sb.WriteString(fmt.Sprintf("input %s {\n", typ.Name))
			for _, field := range typ.InputFields {
				err = printDescription(sb, typ.Description)
				if err != nil {
					return fmt.Errorf("unable to write description for input field %s.%s: %w", typ.Name, field.Name, err)
				}
				sb.WriteString(fmt.Sprintf("\t%s: %s\n", field.Name, introspectionTypeToAstType(field.Type).String()))
			}
			sb.WriteString("}")

		case ast.Interface:
			err = printInterface(sb, typ)
			if err != nil {
				return fmt.Errorf("unable to write interface %s: %w", typ.Name, err)
			}
		default:
			return fmt.Errorf("unsupported type for %s: %s", typ.Name, typ.Kind)
		}
		sb.WriteString("\n")
		sb.WriteString("\n")
	}
	return nil
}

func printDescription(sb *strings.Builder, description string) error {
	if description != "" {
		sb.WriteString(`"""`)
		sb.WriteString("\n")
		sb.WriteString(description)
		sb.WriteString("\n")
		sb.WriteString(`"""`)
		sb.WriteString("\n")
	}
	return nil
}

func printInterface(sb *strings.Builder, typ introspectionTypeDefinition) error {
	if typ.Kind != ast.Interface {
		return fmt.Errorf("cannot print %v as %v", typ.Kind, ast.Interface)
	}

	sb.WriteString(fmt.Sprintf("interface %s {\n", typ.Name))
	for _, field := range typ.Fields {
		err := printDescription(sb, typ.Description)
		if err != nil {
			return fmt.Errorf("unable to write description for field %s: %w", field.Name, err)
		}
		sb.WriteString(fmt.Sprintf("\t%s", field.Name))
		if len(field.Args) > 0 {
			sb.WriteString("(\n")
			for _, arg := range field.Args {
				sb.WriteString(fmt.Sprintf("\t\t%s: %s\n", arg.Name, introspectionTypeToAstType(arg.Type).String()))
			}
			sb.WriteString("\t)")
		}
		sb.WriteString(fmt.Sprintf(": %s\n", introspectionTypeToAstType(field.Type).String()))
	}
	sb.WriteString("}")

	return nil
}
