// SPDX-License-Identifier: MPL-2.0

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package nullable

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
	"time"
)

// TestStruct is a structure used for testing Scan and Value methods.
type TestStruct struct {
	Field string
}

// Scan implements the sql.Scanner interface for TestStruct.
func (t *TestStruct) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan TestStruct")
	}
	t.Field = str
	return nil
}

// Value implements the driver.Valuer interface for TestStruct.
func (t TestStruct) Value() (driver.Value, error) {
	return t.Field, nil
}

func TestNullable_Constructors(t *testing.T) {
	t.Run("NewNullable with value", func(t *testing.T) {
		n := FromValue("test")
		assert.True(t, n.valid)
		assert.Equal(t, "test", n.value)
	})

	t.Run("Null constructor", func(t *testing.T) {
		n := Null[int]()
		assert.False(t, n.valid)
		assert.Equal(t, 0, n.value) // Zero value for int
	})

	t.Run("FromPointer with non-nil pointer", func(t *testing.T) {
		val := 42
		n := FromPointer(&val)
		assert.True(t, n.valid)
		assert.Equal(t, 42, n.value)
	})

	t.Run("FromPointer with nil pointer", func(t *testing.T) {
		var val *string = nil
		n := FromPointer(val)
		assert.False(t, n.valid)
		var zero string
		assert.Equal(t, zero, n.value)
	})
}

func TestNullable_Methods(t *testing.T) {
	t.Run("OrElse when valid is true", func(t *testing.T) {
		n := FromValue("hello")
		result := n.OrElse("default")
		assert.Equal(t, "hello", result)
	})

	t.Run("OrElse when valid is false", func(t *testing.T) {
		n := Null[string]()
		result := n.OrElse("default")
		assert.Equal(t, "default", result)
	})

	t.Run("GetValue", func(t *testing.T) {
		n := FromValue(100)
		assert.Equal(t, 100, n.GetValue())
	})

	t.Run("IsNull when valid is true", func(t *testing.T) {
		n := FromValue(3.14)
		assert.False(t, n.IsNull())
	})

	t.Run("IsNull when valid is false", func(t *testing.T) {
		n := Null[float64]()
		assert.True(t, n.IsNull())
	})

	t.Run("HasValue", func(t *testing.T) {
		n := FromValue(true)
		assert.True(t, n.HasValue())
	})
}

func TestNullable_JSON(t *testing.T) {
	t.Run("MarshalJSON with valid=true", func(t *testing.T) {
		n := FromValue("json test")
		data, err := json.Marshal(n)
		assert.NoError(t, err)
		assert.JSONEq(t, `"json test"`, string(data))
	})

	t.Run("MarshalJSON with valid=false", func(t *testing.T) {
		n := Null[string]()
		data, err := json.Marshal(n)
		assert.NoError(t, err)
		assert.JSONEq(t, `null`, string(data))
	})

	t.Run("UnmarshalJSON with valid value", func(t *testing.T) {
		var n Nullable[int]
		err := json.Unmarshal([]byte(`123`), &n)
		assert.NoError(t, err)
		assert.True(t, n.valid)
		assert.Equal(t, 123, n.value)
	})

	t.Run("UnmarshalJSON with null", func(t *testing.T) {
		var n Nullable[int]
		err := json.Unmarshal([]byte(`null`), &n)
		assert.NoError(t, err)
		assert.False(t, n.valid)
		assert.Equal(t, 0, n.value) // Zero value for int
	})

	t.Run("UnmarshalJSON with invalid data", func(t *testing.T) {
		var n Nullable[int]
		err := json.Unmarshal([]byte(`"invalid"`), &n)
		assert.Error(t, err)
	})
}

func TestNullable_YAML(t *testing.T) {
	t.Run("MarshalYAML with valid=true", func(t *testing.T) {
		n := FromValue("yaml test")
		data, err := yaml.Marshal(n)
		assert.NoError(t, err)
		assert.Equal(t, "yaml test\n", string(data))
		assert.True(t, n.valid)
	})

	t.Run("MarshalYAML with valid=false", func(t *testing.T) {
		n := Null[string]()
		data, err := yaml.Marshal(n)
		assert.NoError(t, err)
		assert.Equal(t, "null\n", string(data))
		assert.False(t, n.valid)
	})

	t.Run("UnmarshalYAML with valid value", func(t *testing.T) {
		var n Nullable[int]
		err := yaml.Unmarshal([]byte(`123`), &n)
		assert.NoError(t, err)
		assert.True(t, n.valid)
		assert.Equal(t, 123, n.value)
	})

	t.Run("UnmarshalYAML with null", func(t *testing.T) {
		var n Nullable[int]
		err := yaml.Unmarshal([]byte(`null`), &n)
		assert.NoError(t, err)
		assert.False(t, n.valid)
		assert.Equal(t, 0, n.value) // Zero value for int
	})

	t.Run("UnmarshalYAML with invalid data", func(t *testing.T) {
		var n Nullable[int]
		err := yaml.Unmarshal([]byte(`"invalid"`), &n)
		assert.Error(t, err)
	})
}

func TestNullable_DatabaseIntegration(t *testing.T) {
	t.Run("Scan with non-nil value", func(t *testing.T) {
		var n Nullable[string]
		err := n.Scan("database test")
		assert.NoError(t, err)
		assert.True(t, n.valid)
		assert.Equal(t, "database test", n.value)
	})

	t.Run("Scan with nil value", func(t *testing.T) {
		var n Nullable[int]
		err := n.Scan(nil)
		assert.NoError(t, err)
		assert.False(t, n.valid)
		assert.Equal(t, 0, n.value) // Zero value for int
	})

	t.Run("Scan with unsupported type", func(t *testing.T) {
		var n Nullable[int]
		err := n.Scan(3.14) // float64 cannot be converted to int directly
		assert.Error(t, err)
		assert.False(t, n.valid)
	})

	t.Run("Value with valid=true", func(t *testing.T) {
		n := FromValue("driver value test")
		val, err := n.Value()
		assert.NoError(t, err)
		assert.Equal(t, "driver value test", val)
	})

	t.Run("Value with valid=false", func(t *testing.T) {
		n := Null[string]()
		val, err := n.Value()
		assert.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("Scan and Value with custom type implementing sql.Scanner and driver.Valuer", func(t *testing.T) {
		var n Nullable[TestStruct]
		err := n.Scan("custom type")
		assert.NoError(t, err)
		assert.True(t, n.valid)
		assert.Equal(t, TestStruct{Field: "custom type"}, n.value)

		val, err := n.Value()
		assert.NoError(t, err)
		assert.Equal(t, "custom type", val)
	})

	t.Run("Value with unsupported struct type", func(t *testing.T) {
		type Unsupported struct {
			A int
		}
		n := FromValue(Unsupported{A: 1})
		val, err := n.Value()
		assert.Error(t, err)
		assert.Nil(t, val)
	})
}

func TestNullable_EdgeCases(t *testing.T) {
	t.Run("Nullable with zero value", func(t *testing.T) {
		n := FromValue(0)
		assert.True(t, n.valid)
		assert.Equal(t, 0, n.value)
	})

	t.Run("Nullable with pointer type", func(t *testing.T) {
		type Person struct {
			Name string
		}
		p := &Person{Name: "Alice"}
		n := FromValue(p)
		assert.True(t, n.valid)
		assert.Equal(t, p, n.value)

		// Marshal to JSON
		data, err := json.Marshal(n)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"Name":"Alice"}`, string(data))

		// Unmarshal from JSON
		var unmarshaled Nullable[Person]
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.True(t, unmarshaled.valid)
		assert.Equal(t, Person{Name: "Alice"}, unmarshaled.value)
	})

	t.Run("Nullable with slice type", func(t *testing.T) {
		n := FromValue([]int{1, 2, 3})
		assert.True(t, n.valid)
		assert.Equal(t, []int{1, 2, 3}, n.value)

		// Marshal to JSON
		data, err := json.Marshal(n)
		assert.NoError(t, err)
		assert.JSONEq(t, `[1,2,3]`, string(data))

		// Unmarshal from JSON
		var unmarshaled Nullable[[]int]
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.True(t, unmarshaled.valid)
		assert.Equal(t, []int{1, 2, 3}, unmarshaled.value)
	})

	t.Run("Nullable with time.Time", func(t *testing.T) {
		now := time.Now()
		n := FromValue(now)
		assert.True(t, n.valid)
		assert.Equal(t, now, n.value)

		// Marshal to JSON
		data, err := json.Marshal(n)
		assert.NoError(t, err)

		// Unmarshal from JSON
		var unmarshaled Nullable[time.Time]
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.True(t, unmarshaled.valid)
		assert.WithinDuration(t, now, unmarshaled.value, time.Second)
	})

	t.Run("Nullable with unsupported type conversion", func(t *testing.T) {
		var n Nullable[string]
		err := n.Scan(3.14) // Attempt to scan float64 into string
		assert.Error(t, err)
		assert.False(t, n.valid)
	})
}

func TestNullable_UnsupportedConversions(t *testing.T) {
	t.Run("convertToType with incompatible types", func(t *testing.T) {
		_, err := convertToType[int]("string")
		assert.Error(t, err)
	})

	t.Run("convertToDriverValue with unsupported type", func(t *testing.T) {
		type Unsupported struct {
			A int
		}
		_, err := convertToDriverValue(Unsupported{A: 1})
		assert.Error(t, err)
	})

	t.Run("convertToType with numeric conversion (disallowed)", func(t *testing.T) {
		_, err := convertToType[int](float64(42.0))
		assert.Error(t, err)

		_, err = convertToType[int](float64(42.5))
		assert.Error(t, err)
	})
}
