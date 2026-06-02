//go:build memforge_debug

package memarch

import (
	"fmt"
	"reflect"
)

func memarchManualTypeValidate[T any](context string) {
	typeValue := reflect.TypeOf((*T)(nil)).Elem()
	visited := make(map[reflect.Type]struct{})
	memarchManualTypeValidateRecursive(typeValue, typeValue.String(), context, visited)
}

func memarchManualTypeValidateRecursive(
	typeValue reflect.Type,
	path string,
	context string,
	visited map[reflect.Type]struct{},
) {
	if typeValue == nil {
		return
	}
	if _, exists := visited[typeValue]; exists {
		return
	}
	visited[typeValue] = struct{}{}

	switch typeValue.Kind() {
	case reflect.Func:
		panic(fmt.Errorf(
			"memarch: manual type validation failed (%s): function-typed field at %s (%s) cannot be manually allocated",
			context,
			path,
			typeValue.String(),
		))

	case reflect.Struct:
		for fieldIndex := 0; fieldIndex < typeValue.NumField(); fieldIndex++ {
			field := typeValue.Field(fieldIndex)
			fieldPath := path + "." + field.Name
			memarchManualTypeValidateRecursive(field.Type, fieldPath, context, visited)
		}

	case reflect.Array:
		memarchManualTypeValidateRecursive(typeValue.Elem(), path+"[]", context, visited)

	case reflect.Ptr, reflect.Slice:
		memarchManualTypeValidateRecursive(typeValue.Elem(), path+"*", context, visited)
	}
}
