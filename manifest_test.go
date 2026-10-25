package cryptodatadownload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestManifestPreservesPublishedContract(t *testing.T) {
	var schema struct {
		Host     string `json:"host"`
		BasePath string `json:"basePath"`
		Paths    map[string]struct {
			Get struct {
				OperationID string                   `json:"operationId"`
				Parameters  []map[string]interface{} `json:"parameters"`
			} `json:"get"`
		} `json:"paths"`
	}
	var manifest []struct {
		Service          string                   `json:"service"`
		Method           string                   `json:"method"`
		Params           string                   `json:"params"`
		Path             string                   `json:"path"`
		HTTPMethod       string                   `json:"httpMethod"`
		OperationID      string                   `json:"operationId"`
		Parameters       []map[string]interface{} `json:"parameters"`
		ParameterFields  map[string]string        `json:"parameterFields"`
		ProductionURL    string                   `json:"productionURL"`
		DocumentationURL string                   `json:"documentationURL"`
		SchemaURL        string                   `json:"schemaURL"`
		HasRaw           bool                     `json:"hasRaw"`
	}
	read := func(path string, target interface{}) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatal(err)
		}
	}
	read("testdata/openapi.json", &schema)
	read("testdata/operations.json", &manifest)
	if len(schema.Paths) != 116 || len(manifest) != len(schema.Paths) {
		t.Fatalf("schema paths=%d, manifest operations=%d", len(schema.Paths), len(manifest))
	}
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	clientValue := reflect.ValueOf(client).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	seen := make(map[string]bool)
	for _, operation := range manifest {
		t.Run(operation.Path, func(t *testing.T) {
			source, exists := schema.Paths[operation.Path]
			if !exists || seen[operation.Path] || operation.HTTPMethod != "GET" || !strings.HasSuffix(operation.Path, "/") {
				t.Fatalf("missing, duplicate or incorrect route: %s", operation.Path)
			}
			seen[operation.Path] = true
			if operation.OperationID != source.Get.OperationID || !reflect.DeepEqual(operation.Parameters, source.Get.Parameters) {
				t.Fatal("operationId, parameters, required fields, defaults or enums differ from the schema")
			}
			base := "https://" + schema.Host
			pointer := strings.ReplaceAll(strings.ReplaceAll(operation.Path, "~", "~0"), "/", "~1")
			if operation.ProductionURL != base+schema.BasePath+operation.Path || operation.DocumentationURL != base+"/" || operation.SchemaURL != base+"/?format=openapi#/paths/"+pointer+"/get" {
				t.Fatal("production or documentation URL differs from the source contract")
			}
			service := clientValue.FieldByName(operation.Service)
			if !service.IsValid() || service.IsNil() {
				t.Fatalf("missing service %s", operation.Service)
			}
			method := service.MethodByName(operation.Method)
			if !method.IsValid() || method.Type().NumIn() != 2 || method.Type().In(0) != contextType || method.Type().In(1).Name() != operation.Params {
				t.Fatalf("missing typed method or query struct: %s.%s", operation.Service, operation.Method)
			}
			if !service.MethodByName(operation.Method + "WithResponse").IsValid() {
				t.Fatal("missing response-metadata companion")
			}
			queryType := method.Type().In(1)
			hasReturn := false
			for _, parameter := range source.Get.Parameters {
				wireName := parameter["name"].(string)
				if wireName == "auth_token" {
					continue
				}
				if wireName == "return" {
					hasReturn = true
					continue
				}
				field, exists := queryType.FieldByName(operation.ParameterFields[wireName])
				if !exists {
					t.Fatalf("parameter %s has no typed query field", wireName)
				}
				required, _ := parameter["required"].(bool)
				if required && field.Type.Kind() == reflect.Pointer || !required && field.Type.Kind() != reflect.Pointer && wireName != "fields" {
					t.Errorf("required/optional semantics changed for %s", wireName)
				}
			}
			if operation.HasRaw != hasReturn || service.MethodByName(operation.Method+"Raw").IsValid() != hasReturn || service.MethodByName(operation.Method+"Stream").IsValid() != hasReturn {
				t.Fatal("raw/stream formats differ from the operation's declared return parameter")
			}
		})
	}
}

func TestAllQueryFieldsAndDownloadVariantsMatchManifest(t *testing.T) {
	data, err := os.ReadFile("testdata/operations.json")
	if err != nil {
		t.Fatal(err)
	}
	var operations []struct {
		Service         string                   `json:"service"`
		Method          string                   `json:"method"`
		Path            string                   `json:"path"`
		Shape           string                   `json:"shape"`
		Parameters      []map[string]interface{} `json:"parameters"`
		ParameterFields map[string]string        `json:"parameterFields"`
		HasRaw          bool                     `json:"hasRaw"`
	}
	if err := json.Unmarshal(data, &operations); err != nil {
		t.Fatal(err)
	}
	for _, operation := range operations {
		t.Run(operation.Service+"."+operation.Method, func(t *testing.T) {
			variants := []string{""}
			if operation.HasRaw {
				variants = append(variants, "Raw", "Stream")
			}
			for _, variant := range variants {
				expected := url.Values{}
				if variant != "" {
					expected.Set("return", "CSV")
				}
				requests := 0
				client := transportTestClient(t, roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					requests++
					if r.URL.Path != "/v1"+operation.Path || !reflect.DeepEqual(r.URL.Query(), expected) || r.Header.Get("Authorization") != "Token test-key" {
						t.Errorf("%s: request=%s, want path=/v1%s query=%v", variant, r.URL.String(), operation.Path, expected)
					}
					body := `{"result":[]}`
					if operation.Shape == "object" {
						body = `{"result":{}}`
					}
					return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
				}))
				method := reflect.ValueOf(client).Elem().FieldByName(operation.Service).MethodByName(operation.Method + variant)
				query := reflect.New(method.Type().In(1)).Elem()
				for _, parameter := range operation.Parameters {
					wireName := parameter["name"].(string)
					if wireName == "auth_token" || wireName == "return" {
						continue
					}
					field := query.FieldByName(operation.ParameterFields[wireName])
					if field.Kind() == reflect.Pointer {
						field.Set(reflect.New(field.Type().Elem()))
						field = field.Elem()
					}
					var value interface{}
					if enum, ok := parameter["enum"].([]interface{}); ok && len(enum) > 0 {
						value = enum[0]
					}
					switch {
					case field.Type() == reflect.TypeOf(time.Time{}):
						field.Set(reflect.ValueOf(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
						expected.Set(wireName, "2024-01-01")
					case field.Kind() == reflect.Slice:
						field.Set(reflect.ValueOf([]string{"symbol", "date"}))
						expected.Set(wireName, "symbol,date")
					case field.Kind() == reflect.String:
						if value == nil {
							value = "BTC /+?&=test"
						}
						field.SetString(value.(string))
						expected.Set(wireName, value.(string))
					case field.Kind() == reflect.Int:
						number := int64(1)
						if value != nil {
							number = int64(value.(float64))
						}
						field.SetInt(number)
						expected.Set(wireName, fmt.Sprint(number))
					case field.Kind() == reflect.Float64:
						field.SetFloat(1.5)
						expected.Set(wireName, "1.5")
					default:
						t.Fatalf("unsupported query field %s", field.Type())
					}
				}
				args := []reflect.Value{reflect.ValueOf(context.Background()), query}
				if variant != "" {
					args = append(args, reflect.ValueOf(FormatCSV))
				}
				result := method.Call(args)
				if failure := result[len(result)-1]; !failure.IsNil() {
					t.Fatalf("%s: %v", variant, failure.Interface())
				}
				if variant == "Stream" {
					if err := result[0].Interface().(*RawResponse).Close(); err != nil {
						t.Fatal(err)
					}
				}
				if requests != 1 {
					t.Fatalf("%s: requests=%d, want 1", variant, requests)
				}
			}
		})
	}
}
