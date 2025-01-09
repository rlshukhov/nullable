// SPDX-License-Identifier: MPL-2.0

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package nullable provides a generic Nullable type for handling nullable values.
// This package supports serialization to JSON and YAML, as well as integration with databases
// through the sql.Scanner and driver.Valuer interfaces.
package nullable

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
)

// ErrUnsupportedConversion occurs when attempting to convert a value to an unsupported type.
var (
	ErrUnsupportedConversion = errors.New("unsupported type conversion")
)

// Nullable represents a nullable value of any type T.
// The value field holds the actual value of type T.
// valid indicates whether the value is set (true) or is null (false).
type Nullable[T any] struct {
	value T
	valid bool
}

// FromValue creates a Nullable with the given value and sets valid to true.
func FromValue[T any](value T) Nullable[T] {
	return Nullable[T]{value: value, valid: true}
}

// Null creates a new Nullable without a value (valid = false).
func Null[T any]() Nullable[T] {
	return Nullable[T]{valid: false}
}

// FromPointer creates a Nullable from a pointer. If the pointer is nil, valid is set to false.
func FromPointer[T any](value *T) Nullable[T] {
	if value == nil {
		return Nullable[T]{valid: false}
	}
	return Nullable[T]{value: *value, valid: true}
}

// OrElse returns the value if valid is true; otherwise, it returns the provided defaultVal.
func (n Nullable[T]) OrElse(defaultVal T) T {
	if n.valid {
		return n.value
	}
	return defaultVal
}

// GetValue returns the actual value T.
func (n Nullable[T]) GetValue() T {
	return n.value
}

// IsNull checks if the value is null (valid = false).
func (n Nullable[T]) IsNull() bool {
	return !n.valid
}

// HasValue checks if the value is not null (valid = true).
func (n Nullable[T]) HasValue() bool {
	return n.valid
}

// Scan implements the sql.Scanner interface for Nullable, allowing it to be used in database operations.
func (n *Nullable[T]) Scan(value any) error {
	if value == nil {
		n.value = zeroValue[T]()
		n.valid = false
		return nil
	}

	// Check if *T implements sql.Scanner
	if scanner, ok := any(&n.value).(sql.Scanner); ok {
		err := scanner.Scan(value)
		if err != nil {
			n.valid = false
			return err
		}
		n.valid = true
		return nil
	}

	// If T does not implement sql.Scanner, attempt type conversion
	var err error
	n.value, err = convertToType[T](value)
	if err != nil {
		n.valid = false
		return err
	}
	n.valid = true
	return nil
}

// Value implements the driver.Valuer interface for Nullable, allowing it to be used in database operations.
func (n Nullable[T]) Value() (driver.Value, error) {
	if !n.valid {
		return nil, nil
	}

	// Check if T implements driver.Valuer
	if valuer, ok := any(n.value).(driver.Valuer); ok {
		return valuer.Value()
	}

	return convertToDriverValue(n.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface for Nullable.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.valid = false
		n.value = zeroValue[T]()
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		n.valid = false
		return err
	}

	n.value = value
	n.valid = true
	return nil
}

// MarshalJSON implements the json.Marshaler interface for Nullable.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.valid {
		return json.Marshal(nil)
	}

	return json.Marshal(n.value)
}

// UnmarshalYAML implements the unmarshaling of YAML data.
func (n *Nullable[T]) UnmarshalYAML(unmarshal func(any) error) error {
	var value *T
	if err := unmarshal(&value); err != nil {
		n.valid = false
		return err
	}

	if value == nil {
		n.valid = false
		n.value = zeroValue[T]()
	} else {
		n.value = *value
		n.valid = true
	}

	return nil
}

// MarshalYAML implements the marshaling of YAML data.
func (n Nullable[T]) MarshalYAML() (any, error) {
	if !n.valid {
		return nil, nil
	}
	return n.value, nil
}

// convertToDriverValue converts a value to driver.Value for use with databases.
func convertToDriverValue(v any) (driver.Value, error) {
	if valuer, ok := v.(driver.Valuer); ok {
		return valuer.Value()
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			return nil, nil
		}
		return convertToDriverValue(rv.Elem().Interface())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return int64(rv.Uint()), nil

	case reflect.Uint64:
		u64 := rv.Uint()
		if u64 >= 1<<63 {
			return nil, fmt.Errorf("uint64 values with high bit set are not supported")
		}
		return int64(u64), nil

	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil

	case reflect.Bool:
		return rv.Bool(), nil

	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return rv.Bytes(), nil
		}
		return nil, fmt.Errorf("unsupported slice type: %s", rv.Type().Elem().Kind())

	case reflect.String:
		return rv.String(), nil

	case reflect.Struct:
		if t, ok := v.(time.Time); ok {
			return t, nil
		}
		return nil, fmt.Errorf("unsupported struct type: %s", rv.Type())

	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

// convertToType attempts to convert a value to type T.
// In this implementation, conversion between different types, even if numeric, is disallowed to ensure strict typing.
func convertToType[T any](value any) (T, error) {
	var zero T
	if value == nil {
		return zero, nil
	}

	valueType := reflect.TypeOf(value)
	targetType := reflect.TypeOf(zero)
	if valueType == targetType {
		return value.(T), nil
	}

	return zero, ErrUnsupportedConversion
}

// zeroValue returns the zero value for type T.
func zeroValue[T any]() T {
	var zero T
	return zero
}
