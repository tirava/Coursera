package main

import (
	"errors"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {

	outPtr := reflect.ValueOf(out)
	if outPtr.Kind() != reflect.Ptr {
		return errors.New("need Ptr only")
	}
	outElem := outPtr.Elem()

	switch outElem.Kind() {

	case reflect.Struct:
		mapI, ok := data.(map[string]interface{})
		if !ok {
			return errors.New("need map interface")
		}
		for k, v := range mapI {
			o := reflect.New(outElem.FieldByName(k).Type()).Interface()
			err := i2s(v, o)
			if err != nil {
				return errors.New("error convert to struct")
			}
			outElem.FieldByName(k).Set(reflect.ValueOf(o).Elem())
		}

	case reflect.Slice:
		sliceI, ok := data.([]interface{})
		if !ok {
			return errors.New("need slice")
		}
		elems := reflect.New(outElem.Type()).Elem()
		for _, s := range sliceI {
			o := reflect.New(outElem.Type().Elem()).Interface()
			err := i2s(s, o)
			if err != nil {
				return errors.New("error convert to struct")
			}
			elems = reflect.Append(elems, reflect.ValueOf(o).Elem())
		}
		outElem.Set(elems)

	case reflect.Int:
		val, ok := data.(float64)
		if !ok {
			return errors.New("need int")
		}
		outElem.SetInt(int64(val))

	case reflect.String:
		val, ok := data.(string)
		if !ok {
			return errors.New("need string")
		}
		outElem.SetString(val)

	case reflect.Bool:
		val, ok := data.(bool)
		if !ok {
			return errors.New("need bool")
		}
		outElem.SetBool(val)

	default:
		return errors.New("unknown error")

	}

	return nil
}
