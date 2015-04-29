package merge

import "reflect"

// Merge will take two data sets and merge them together - returning a new data set
func Merge(base, override interface{}) interface{} {
	//reflect and recurse
	b := reflect.ValueOf(base)
	o := reflect.ValueOf(override)
	ret := mergeRecursive(b, o)

	return ret.Interface()
}

func mergeRecursive(base, override reflect.Value) reflect.Value {
	var result reflect.Value

	switch base.Kind() {
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		// Pointers and Interfaces should just be unwrapped and recursed through
		result = mergeRecursive(base.Elem(), override.Elem())

	case reflect.Struct:
		// For structs we loop over fields and recurse
		// setup our result struct
		result = reflect.New(base.Type())
		for i, n := 0, base.NumField(); i < n; i++ {
			// get the merged value of each field
			newVal := mergeRecursive(base.Field(i), override.Field(i))

			//attempt to set that merged value on our result struct
			if result.Elem().Field(i).CanSet() {
				result.Elem().Field(i).Set(newVal)
			}
		}

	case reflect.Map:
		// For Maps we copy the base data, and then replace it with merged data
		// We use two for loops to make sure all map keys from base and all keys from
		// override exist in the result just in case one of the maps is sparse.

		result = reflect.MakeMap(base.Type())
		// Copy from base first
		for _, key := range base.MapKeys() {
			result.SetMapIndex(key, base.MapIndex(key))
		}

		// Override with values from override if they exist
		for _, key := range override.MapKeys() {
			overrideVal := override.MapIndex(key)
			baseVal := base.MapIndex(key)
			if !overrideVal.IsValid() {
				continue
			}

			// Merge the values and set in the result
			newVal := mergeRecursive(baseVal, overrideVal)
			if newVal.Kind() == reflect.Ptr {
				result.SetMapIndex(key, newVal.Elem())

			} else {
				result.SetMapIndex(key, newVal)
			}
		}

	default:
		// These are all generic types
		// override will be taken for generic types if it is set
		if isEmptyValue(override) {
			result = base
		} else {
			result = override
		}
	}
	return result
}

// Copied From http://golang.org/src/encoding/json/encode.go
// Lines 280 - 296
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}