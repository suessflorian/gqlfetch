package gqlfetch

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

func strPtr(s string) *string { return &s }

func Test_printInterface(t *testing.T) {
	type args struct {
		sb  strings.Builder
		typ introspectionTypeDefinition
	}
	tests := map[string]struct {
		args   args
		expect string
	}{
		"zero argument object field": {
			args: args{
				sb: strings.Builder{},
				typ: introspectionTypeDefinition{
					Kind: ast.Interface,
					Name: "MyInterface",
					Fields: []introspectedTypeField{
						{
							Name: "myInterfaceField",
							Type: &introspectedType{
								Name: strPtr("myResponseObject"),
								Kind: OBJECT,
							},
							Args: nil,
						},
					},
				},
			},
			expect: "interface MyInterface {\n\tmyInterfaceField: myResponseObject\n}",
		},
		"one argument object field": {
			args: args{
				sb: strings.Builder{},
				typ: introspectionTypeDefinition{
					Kind: ast.Interface,
					Name: "MyInterface",
					Fields: []introspectedTypeField{
						{
							Name: "myInterfaceField",
							Type: &introspectedType{
								Name: strPtr("myResponseObject"),
								Kind: OBJECT,
							},
							Args: []introspectionInputField{
								{
									Name:        "argOne",
									Description: "",
									Type: &introspectedType{
										Name: strPtr("Int"),
										Kind: NON_NULL,
									},
								},
							},
						},
					},
				},
			},
			expect: `interface MyInterface {
	myInterfaceField(
		argOne: Int
	): myResponseObject
}`,
		},
		"many argument object field": {
			args: args{
				sb: strings.Builder{},
				typ: introspectionTypeDefinition{
					Kind:        ast.Interface,
					Name:        "MyInterface",
					Description: "",
					Fields: []introspectedTypeField{
						{
							Name: "myInterfaceField",
							Type: &introspectedType{
								Name: strPtr("myResponseObject"),
								Kind: OBJECT,
							},
							Args: []introspectionInputField{
								{
									Name: "argOne",
									Type: &introspectedType{
										Name: strPtr("Int"),
										Kind: NON_NULL,
									},
								},
								{
									Name: "argTwo",
									Type: &introspectedType{
										Name: strPtr("Int"),
										Kind: NON_NULL,
									},
								},
							},
						},
					},
				},
			},
			expect: `interface MyInterface {
	myInterfaceField(
		argOne: Int
		argTwo: Int
	): myResponseObject
}`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			printInterface(&tt.args.sb, tt.args.typ)
			got := tt.args.sb.String()
			if got != tt.expect {
				t.Errorf("printing ast.Interface expect: %v got: %v", tt.expect, got)
			}
		})
	}
}

func Test_printTypes(t *testing.T) {
	tests := map[string]struct {
		typ    introspectionTypeDefinition
		expect string
	}{
		"object with deprecated field": {
			typ: introspectionTypeDefinition{
				Kind: ast.Object,
				Name: "User",
				Fields: []introspectedTypeField{
					{
						Name: "oldField",
						Type: &introspectedType{
							Name: strPtr("String"),
							Kind: OBJECT,
						},
						IsDeprecated:      true,
						DeprecationReason: "Use newField instead",
					},
					{
						Name: "newField",
						Type: &introspectedType{
							Name: strPtr("String"),
							Kind: OBJECT,
						},
					},
				},
			},
			expect: `type User {
	oldField: String @deprecated(reason: "Use newField instead")
	newField: String
}

`,
		},
		"enum with deprecated value": {
			typ: introspectionTypeDefinition{
				Kind: ast.Enum,
				Name: "Status",
				EnumValues: json.RawMessage(`[
					{
						"name": "OLD",
						"description": "",
						"isDeprecated": true,
						"deprecationReason": "Use NEW instead"
					},
					{
						"name": "NEW",
						"description": "",
						"isDeprecated": false,
						"deprecationReason": null
					}
				]`),
			},
			expect: `enum Status {
	OLD @deprecated(reason: "Use NEW instead")
	NEW
}

`,
		},
		"object with deprecated field without reason": {
			typ: introspectionTypeDefinition{
				Kind: ast.Object,
				Name: "User",
				Fields: []introspectedTypeField{
					{
						Name: "oldField",
						Type: &introspectedType{
							Name: strPtr("String"),
							Kind: OBJECT,
						},
						IsDeprecated: true,
					},
				},
			},
			expect: `type User {
	oldField: String @deprecated
}

`,
		},
		"enum with deprecated value without reason": {
			typ: introspectionTypeDefinition{
				Kind: ast.Enum,
				Name: "Status",
				EnumValues: json.RawMessage(`[
					{
						"name": "OLD",
						"description": "",
						"isDeprecated": true,
						"deprecationReason": null
					}
				]`),
			},
			expect: `enum Status {
	OLD @deprecated
}

`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sb := &strings.Builder{}
			err := printTypes(sb, []introspectionTypeDefinition{tt.typ}, false)
			if err != nil {
				t.Errorf("printTypes() error = %v", err)
				return
			}
			got := sb.String()
			if got != tt.expect {
				t.Errorf("printTypes() = %v, want %v", got, tt.expect)
			}
		})
	}
}
