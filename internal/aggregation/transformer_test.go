package aggregation

import (
	"testing"
)

func TestResponseTransformer(t *testing.T) {
	rt := NewResponseTransformer()

	testData := map[string]interface{}{
		"id":       "123",
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "secret",
		"role":     "admin",
	}

	t.Run("Pick", func(t *testing.T) {
		result, err := rt.Pick(testData, []string{"id", "name"})
		if err != nil {
			t.Fatalf("Pick failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(result))
		}

		if result["id"] != "123" {
			t.Error("id field incorrect")
		}

		if result["name"] != "John Doe" {
			t.Error("name field incorrect")
		}

		if _, exists := result["password"]; exists {
			t.Error("password should not be in result")
		}
	})

	t.Run("Omit", func(t *testing.T) {
		result, err := rt.Omit(testData, []string{"password"})
		if err != nil {
			t.Fatalf("Omit failed: %v", err)
		}

		if _, exists := result["password"]; exists {
			t.Error("password should be omitted")
		}

		if result["id"] != "123" {
			t.Error("id should still be present")
		}
	})

	t.Run("Sanitize", func(t *testing.T) {
		result, err := rt.Sanitize(testData)
		if err != nil {
			t.Fatalf("Sanitize failed: %v", err)
		}

		if _, exists := result["password"]; exists {
			t.Error("password should be sanitized")
		}

		if result["id"] != "123" {
			t.Error("non-sensitive fields should remain")
		}
	})

	t.Run("Rename", func(t *testing.T) {
		mapping := map[string]string{
			"id":   "userId",
			"name": "fullName",
		}

		result, err := rt.Rename(testData, mapping)
		if err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		if _, exists := result["id"]; exists {
			t.Error("Old key 'id' should not exist")
		}

		if result["userId"] != "123" {
			t.Error("userId should have value from id")
		}

		if result["fullName"] != "John Doe" {
			t.Error("fullName should have value from name")
		}
	})

	t.Run("Flatten", func(t *testing.T) {
		nested := map[string]interface{}{
			"user": map[string]interface{}{
				"id":   "123",
				"name": "John",
				"address": map[string]interface{}{
					"city": "NYC",
				},
			},
		}

		result, err := rt.Flatten(nested, "")
		if err != nil {
			t.Fatalf("Flatten failed: %v", err)
		}

		if result["user.id"] != "123" {
			t.Error("Flattened key user.id incorrect")
		}

		if result["user.address.city"] != "NYC" {
			t.Error("Deeply nested key incorrect")
		}
	})

	t.Run("Merge", func(t *testing.T) {
		map1 := map[string]interface{}{"a": 1, "b": 2}
		map2 := map[string]interface{}{"c": 3, "b": 99}
		map3 := map[string]interface{}{"d": 4}

		result := rt.Merge(map1, map2, map3)

		if result["a"] != 1 {
			t.Error("Value from map1 incorrect")
		}

		if result["b"] != 99 {
			t.Error("Value should be overwritten by map2")
		}

		if result["c"] != 3 {
			t.Error("Value from map2 incorrect")
		}

		if result["d"] != 4 {
			t.Error("Value from map3 incorrect")
		}
	})

	t.Run("AddComputedFields", func(t *testing.T) {
		data := map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"age":       30,
		}

		computed := map[string]func(map[string]interface{}) interface{}{
			"fullName": func(d map[string]interface{}) interface{} {
				return d["firstName"].(string) + " " + d["lastName"].(string)
			},
			"isAdult": func(d map[string]interface{}) interface{} {
				return d["age"].(int) >= 18
			},
		}

		result, err := rt.AddComputedFields(data, computed)
		if err != nil {
			t.Fatalf("AddComputedFields failed: %v", err)
		}

		if result["fullName"] != "John Doe" {
			t.Error("Computed fullName incorrect")
		}

		if result["isAdult"] != true {
			t.Error("Computed isAdult incorrect")
		}
	})

	t.Run("FilterNull", func(t *testing.T) {
		data := map[string]interface{}{
			"a": "value",
			"b": nil,
			"c": "another",
			"d": nil,
		}

		result := rt.FilterNull(data)

		if len(result) != 2 {
			t.Errorf("Expected 2 non-null values, got %d", len(result))
		}

		if _, exists := result["b"]; exists {
			t.Error("Null value b should be filtered")
		}
	})

	t.Run("ToJSON_FromJSON", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "test",
			"value": 123,
		}

		// Convert to JSON
		jsonStr, err := rt.ToJSON(data)
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}

		if jsonStr == "" {
			t.Error("JSON string should not be empty")
		}

		// Convert back from JSON
		parsed, err := rt.FromJSON(jsonStr)
		if err != nil {
			t.Fatalf("FromJSON failed: %v", err)
		}

		if parsed["name"] != "test" {
			t.Error("Parsed name incorrect")
		}

		// Note: JSON numbers are float64
		if parsed["value"].(float64) != 123.0 {
			t.Error("Parsed value incorrect")
		}
	})

	t.Run("SelectFields", func(t *testing.T) {
		data := map[string]interface{}{
			"id":    "123",
			"name":  "John",
			"email": "john@example.com",
			"role":  "admin",
		}

		result, err := rt.SelectFields(data, "id,name")
		if err != nil {
			t.Fatalf("SelectFields failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result should be a map")
		}

		if len(resultMap) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(resultMap))
		}
	})
}
