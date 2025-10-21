package aggregation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ResponseTransformer provides utilities for transforming API responses
type ResponseTransformer struct{}

// NewResponseTransformer creates a new response transformer
func NewResponseTransformer() *ResponseTransformer {
	return &ResponseTransformer{}
}

// Pick selects specific fields from a map or struct
func (rt *ResponseTransformer) Pick(data interface{}, fields []string) (map[string]interface{}, error) {
	// Convert to map
	dataMap, err := toMap(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, field := range fields {
		if value, ok := dataMap[field]; ok {
			result[field] = value
		}
	}

	return result, nil
}

// Omit removes specific fields from a map or struct
func (rt *ResponseTransformer) Omit(data interface{}, fields []string) (map[string]interface{}, error) {
	// Convert to map
	dataMap, err := toMap(data)
	if err != nil {
		return nil, err
	}

	// Create exclusion set
	exclude := make(map[string]bool)
	for _, field := range fields {
		exclude[field] = true
	}

	result := make(map[string]interface{})
	for key, value := range dataMap {
		if !exclude[key] {
			result[key] = value
		}
	}

	return result, nil
}

// Sanitize removes common sensitive fields
func (rt *ResponseTransformer) Sanitize(data interface{}) (map[string]interface{}, error) {
	sensitiveFields := []string{
		"password", "passwordHash", "secret", "token",
		"apiKey", "privateKey", "ssn", "creditCard",
		"cvv", "pin", "accessToken", "refreshToken",
	}

	return rt.Omit(data, sensitiveFields)
}

// Rename renames fields in a map
func (rt *ResponseTransformer) Rename(data interface{}, mapping map[string]string) (map[string]interface{}, error) {
	dataMap, err := toMap(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for key, value := range dataMap {
		if newKey, ok := mapping[key]; ok {
			result[newKey] = value
		} else {
			result[key] = value
		}
	}

	return result, nil
}

// Flatten flattens nested objects with dot notation
func (rt *ResponseTransformer) Flatten(data interface{}, prefix string) (map[string]interface{}, error) {
	dataMap, err := toMap(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	rt.flattenRecursive(dataMap, prefix, result)

	return result, nil
}

// flattenRecursive recursively flattens nested maps
func (rt *ResponseTransformer) flattenRecursive(data map[string]interface{}, prefix string, result map[string]interface{}) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// Check if value is a map
		if valueMap, ok := value.(map[string]interface{}); ok {
			rt.flattenRecursive(valueMap, fullKey, result)
		} else {
			result[fullKey] = value
		}
	}
}

// Merge combines multiple maps into one
func (rt *ResponseTransformer) Merge(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, m := range maps {
		for key, value := range m {
			result[key] = value
		}
	}

	return result
}

// TransformArray applies a transformation function to each item in a slice
func (rt *ResponseTransformer) TransformArray(data interface{}, transformFn func(interface{}) (interface{}, error)) ([]interface{}, error) {
	// Use reflection to handle different slice types
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("data is not a slice or array")
	}

	result := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		transformed, err := transformFn(item)
		if err != nil {
			return nil, fmt.Errorf("failed to transform item %d: %w", i, err)
		}
		result[i] = transformed
	}

	return result, nil
}

// AddComputedFields adds computed fields to data
func (rt *ResponseTransformer) AddComputedFields(data interface{}, computedFields map[string]func(map[string]interface{}) interface{}) (map[string]interface{}, error) {
	dataMap, err := toMap(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for key, value := range dataMap {
		result[key] = value
	}

	for fieldName, computeFn := range computedFields {
		result[fieldName] = computeFn(dataMap)
	}

	return result, nil
}

// MapKeys applies a function to all keys
func (rt *ResponseTransformer) MapKeys(data map[string]interface{}, mapFn func(string) string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		newKey := mapFn(key)
		result[newKey] = value
	}
	return result
}

// MapValues applies a function to all values
func (rt *ResponseTransformer) MapValues(data map[string]interface{}, mapFn func(interface{}) interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		result[key] = mapFn(value)
	}
	return result
}

// toMap converts various types to map[string]interface{}
func toMap(data interface{}) (map[string]interface{}, error) {
	// If already a map, return it
	if m, ok := data.(map[string]interface{}); ok {
		return m, nil
	}

	// Try JSON marshal/unmarshal for structs
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return result, nil
}

// SelectFields selects fields from data based on field selector string
func (rt *ResponseTransformer) SelectFields(data interface{}, fields string) (interface{}, error) {
	if fields == "" {
		return data, nil
	}

	fieldList := strings.Split(fields, ",")
	for i, field := range fieldList {
		fieldList[i] = strings.TrimSpace(field)
	}

	return rt.Pick(data, fieldList)
}

// FilterNull removes null/nil values from a map
func (rt *ResponseTransformer) FilterNull(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		if value != nil {
			result[key] = value
		}
	}
	return result
}

// ToJSON converts data to JSON string
func (rt *ResponseTransformer) ToJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON parses JSON string to map
func (rt *ResponseTransformer) FromJSON(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return result, nil
}
