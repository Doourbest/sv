package utils

import "reflect"

func HasValue(arr interface{},e interface{}) bool {
	arrV := reflect.ValueOf(arr)
	if arrV.Kind()==reflect.Ptr && !arrV.IsNil() && (arrV.Elem().Kind()==reflect.Slice || arrV.Elem().Kind()==reflect.Array) {
		arrV = arrV.Elem()
	}
	if arrV.Kind()==reflect.Slice || arrV.Kind()==reflect.Array {
		for i:=0; i<arrV.Len(); i+=1 {
			if reflect.DeepEqual(e,arrV.Index(i).Interface()) {
				return true;
			}
		}
	}
	return false;
}

