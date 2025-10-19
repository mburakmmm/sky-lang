package skylib

import (
	"fmt"
	"reflect"
)

// TypeName returns type name of value
func ReflectTypeName(obj interface{}) string {
	if obj == nil {
		return "nil"
	}
	return reflect.TypeOf(obj).String()
}

// Fields returns field names of struct
func ReflectFields(obj interface{}) []string {
	if obj == nil {
		return []string{}
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return []string{}
	}

	typ := val.Type()
	fields := make([]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		fields[i] = typ.Field(i).Name
	}

	return fields
}

// GetAttr gets attribute value
func ReflectGetAttr(obj interface{}, name string) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot get attr of nil")
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct")
	}

	field := val.FieldByName(name)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found", name)
	}

	return field.Interface(), nil
}

// SetAttr sets attribute value
func ReflectSetAttr(obj interface{}, name string, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("cannot set attr of nil")
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct")
	}

	field := val.FieldByName(name)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", name)
	}

	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", name)
	}

	fieldValue := reflect.ValueOf(value)
	if !fieldValue.Type().AssignableTo(field.Type()) {
		return fmt.Errorf("type mismatch")
	}

	field.Set(fieldValue)
	return nil
}
