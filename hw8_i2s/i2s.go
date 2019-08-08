package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {

	//elems := reflect.ValueOf(data).MapRange()
	//elemsType := elems.Type()
	//fmt.Println().Println(elems)

	//val := reflect.ValueOf(data)
	//fmt.Println(data.(type))
	//fmt.Println("VALUE = ", val)
	//fmt.Println("KIND = ", val.Kind())
	//switch data.(type) {
	//default:
	//case map[string]interface{}:
	//	println("mmmmmmmmmmmmmmmmmmmmmmmmmmmmm")
	//}

	//switch val.Kind() {
	//case reflect.Map:
	//	iter := reflect.ValueOf(data).MapRange()
	//	for iter.Next() {
	//		k := iter.Key()
	//		v := iter.Value()
	//		//t := v.Type()
	//		fmt.Println("kkk:", k, "vvv:", v, "ttt:")
	//		v.Convert(reflect.Int)
	//	}
	//for _, e := range val.MapKeys() {
	//	v := val.MapIndex(e)
	//	switch t := v.Interface().(type) {
	//	case int:
	//		fmt.Println(e, t)
	//	case string:
	//		fmt.Println(e, t)
	//	case bool:
	//		fmt.Println(e, t)
	//	default:
	//		fmt.Println("not found")
	//
	//	}
	//}
	//println("mapppppppppp")
	//case reflect.Slice:
	//	//println("sliceeeeeeeeeeeeee")
	//default:
	//	//println("otherssssssssssssssssssss")
	//}

	result, _ := json.Marshal(data)
	//fmt.Println("json:", string(result))
	json.Unmarshal(result, out)

	fmt.Println(reflect.ValueOf(out).Type())

	//return errors.New("111")
	return nil
}

//func CreateMap(key, elem reflect.Type) reflect.Value {
//	var mapType reflect.Type
//	mapType = reflect.MapOf(key, elem)
//	return reflect.MakeMap(mapType)
//}

/*var jsonStr = `[
	{"id": 17, "username": "iivan", "phone": 0},
	{"id": "17", "address": "none", "company": "Mail.ru"}
]`*/

/*func main() {
	data := []byte(jsonStr)

	var user1 interface{}
	json.Unmarshal(data, &user1)
	fmt.Printf("unpacked in empty interface:\n%#v\n\n", user1)

	user2 := map[string]interface{}{
		"id":       42,
		"username": "rvasily",
	}
	var user2i interface{} = user2
	result, _ := json.Marshal(user2i)
	fmt.Printf("json string from map:\n %s\n", string(result))
}*/
