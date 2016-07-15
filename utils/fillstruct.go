
package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)


func FillStruct(m map[string]interface{}, s interface{}) error {
	for k, v := range m {
		err := setField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

var intTypes = [...]reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
}
var uintTypes = [...]reflect.Kind{
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
}

func setField(obj interface{}, k string, v interface{}) error {

	k = UcFirst(k)

	objVal := reflect.ValueOf(obj).Elem()
	fv := objVal.FieldByName(k)

	if !fv.IsValid() {
		return fmt.Errorf("No such field: [%s] in obj", k)
	}

	if !fv.CanSet() {
		return fmt.Errorf("Cannot set [%s] field value", k)
	}

	val := reflect.ValueOf(v)


	if fv.Type() != val.Type() {

		if m,ok := v.(map[string]interface{}); ok {

			// if field value is a pointer to struct, adjust the fv to the value the pointer points to
			if fv.Kind()==reflect.Ptr && fv.Type().Elem().Kind() == reflect.Struct {
				// if the pointer is nil, create the struct
				if fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				fv = fv.Elem()
			}

			// if field value is struct
			if fv.Kind() == reflect.Struct {
				return FillStruct(m, fv.Addr().Interface())
			}

		}

		if val.Kind()==reflect.String {
			s := strings.TrimSpace(val.Interface().(string)) // trim space
			if fv.Kind()==reflect.Bool {
				b,err := strconv.ParseBool(s)
				if err != nil {
					return fmt.Errorf("Field[%s] parse bool from string [%s] failed, %s", k, s, err.Error())
				}
				fv.Set(reflect.ValueOf(b))
				return nil
			}
			if HasValue(intTypes, fv.Kind()) {
				bits := fv.Type().Bits()
				i,err := strconv.ParseInt(s,0,bits)
				if err!=nil {
					return fmt.Errorf("Field[%s] parse [%s] from string [%s] failed, %s", k, fv.Kind().String(), s, err.Error())
				}
				fv.SetInt(i)
				return nil
			}
			if HasValue(uintTypes, fv.Kind()) {
				bits := fv.Type().Bits()
				u,err := strconv.ParseUint(s,0,bits)
				if err!=nil {
					return fmt.Errorf("Field[%s] parse [%s] from string [%s] failed, %s", k, fv.Kind().String(), s, err.Error())
				}
				fv.SetUint(u)
				return nil
			}
		}

		return fmt.Errorf("Field [%s] Provided value type [%s] didn't match struct field type [%s]" , k, val.Type().String(), fv.Type().String())
	}

	fv.Set(val)
	return nil

}

