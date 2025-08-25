package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// OutputData formats and prints data according to the specified format
func OutputData(data interface{}, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return outputJSON(data)
	case "yaml":
		return outputYAML(data)
	case "table":
		return outputTable(data)
	default:
		return outputJSON(data)
	}
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

func outputTable(data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		return outputSliceAsTable(v)
	case reflect.Struct:
		return outputStructAsTable(v)
	default:
		// Fallback to JSON for unsupported types
		return outputJSON(data)
	}
}

func outputSliceAsTable(v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Println("No data found")
		return nil
	}

	// Get the first element to determine structure
	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}

	if first.Kind() != reflect.Struct {
		// For non-struct slices, just print each item
		for i := 0; i < v.Len(); i++ {
			fmt.Println(v.Index(i).Interface())
		}
		return nil
	}

	// Get field names from the first struct
	t := first.Type()
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			headers = append(headers, strings.ToUpper(field.Name))
		}
	}

	// Print headers
	fmt.Println(strings.Join(headers, "\t"))
	fmt.Println(strings.Repeat("-", len(strings.Join(headers, "\t"))))

	// Print data rows
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		var values []string
		for j := 0; j < item.NumField(); j++ {
			field := item.Type().Field(j)
			if field.IsExported() {
				value := item.Field(j)
				values = append(values, fmt.Sprintf("%v", value.Interface()))
			}
		}
		fmt.Println(strings.Join(values, "\t"))
	}

	return nil
}

func outputStructAsTable(v reflect.Value) error {
	t := v.Type()

	fmt.Println("FIELD\tVALUE")
	fmt.Println("-----\t-----")

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			value := v.Field(i)
			fmt.Printf("%s\t%v\n", strings.ToUpper(field.Name), value.Interface())
		}
	}

	return nil
}
