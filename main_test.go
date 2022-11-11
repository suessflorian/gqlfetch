package gqlfetch

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/vektah/gqlparser/ast"
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
