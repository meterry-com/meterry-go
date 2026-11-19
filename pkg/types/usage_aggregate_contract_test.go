package types_test

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"testing"
)

// Compare against a reference contract when updating the request model:
// METERRY_AGGREGATE_CONTRACT_FILE=/path/to/usage_aggregate.go go test ./...
func TestUsageAggregateQueryRequestMatchesReference(t *testing.T) {
	referenceFile := os.Getenv("METERRY_AGGREGATE_CONTRACT_FILE")
	if referenceFile == "" {
		t.Skip("set METERRY_AGGREGATE_CONTRACT_FILE to compare against a reference request contract")
	}
	type fieldContract struct {
		Name      string
		Type      string
		JSON      string
		Anonymous bool
	}
	readRequest := func(path string) []fieldContract {
		t.Helper()
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("read aggregate contract %s: %v", path, err)
		}
		for _, declaration := range file.Decls {
			declaration, ok := declaration.(*ast.GenDecl)
			if !ok || declaration.Tok != token.TYPE {
				continue
			}
			for _, spec := range declaration.Specs {
				spec := spec.(*ast.TypeSpec)
				if spec.Name.Name != "UsageAggregateQueryRequest" {
					continue
				}
				request, ok := spec.Type.(*ast.StructType)
				if !ok {
					t.Fatalf("%s: UsageAggregateQueryRequest must be a struct", path)
				}
				var fields []fieldContract
				for _, field := range request.Fields.List {
					var typeName bytes.Buffer
					if err := format.Node(&typeName, fset, field.Type); err != nil {
						t.Fatal(err)
					}
					contract := fieldContract{Type: typeName.String(), Anonymous: len(field.Names) == 0}
					if field.Tag != nil {
						tag, err := strconv.Unquote(field.Tag.Value)
						if err != nil {
							t.Fatal(err)
						}
						contract.JSON = reflect.StructTag(tag).Get("json")
					}
					if contract.Anonymous {
						fields = append(fields, contract)
					}
					for _, name := range field.Names {
						contract.Name = name.Name
						fields = append(fields, contract)
					}
				}
				return fields
			}
		}
		t.Fatalf("%s: UsageAggregateQueryRequest was not found", path)
		return nil
	}
	want := readRequest(referenceFile)
	got := readRequest("usage_aggregate.go")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("aggregate request differs from the reference contract; synchronize field names, Go types, and JSON tags (including omitempty)\nSDK:       %+v\nreference: %+v", got, want)
	}
}
