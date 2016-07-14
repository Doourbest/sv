package sv

import (
	"reflect"
	"unicode"
	"github.com/gin-gonic/gin"
)

// 自动路由到 Struct 的所有方法，如果定义 Test：
//	type Test struct{
// 		A int
// 	}
// 	func (*Test) Foo(ctx *gin.Context) {
// 		ctx.JSON(200, gin.H{ "message": "pong", })
//	}
// 那么 
//	GinRouteStructMethods(r,&Test{})
// 相当于
//	t = &Text{}
// 	r.GET( "/test/foo",func(ctx *gin.Context) { t.Foo(ctx)})
// 	r.GET( "/Test/Foo",func(ctx *gin.Context) { t.Foo(ctx)})
// 	r.POST("/test/foo",func(ctx *gin.Context) { t.Foo(ctx)})
// 	r.POST("/Test/Foo",func(ctx *gin.Context) { t.Foo(ctx)})
func GinRouteStructMethods(r *gin.Engine, controller interface{}) {

	t := reflect.TypeOf(controller)
	v := reflect.ValueOf(controller)

	for i:=0; i<t.NumMethod(); i=i+1 {

		m := t.Method(i)
		var dummyCtx *gin.Context = nil

		// filter unexported method
		// PkgPath is empty for exported method
		if m.PkgPath!="" {
			return
		}
		// number of parameters
		if m.Type.NumIn()!=2 {
			continue
		}
		if m.Type.In(0) != reflect.TypeOf(controller) {
			continue
		}
		if m.Type.In(1) != reflect.TypeOf(dummyCtx) {
			continue
		}

		mv := v.Method(i)

		callback := func (ctx *gin.Context){
			mv.Call([]reflect.Value{reflect.ValueOf(ctx)})
		}

		cName := ""
		lcName := ""
		if t.Kind()==reflect.Ptr {
			cName = t.Elem().Name()
		} else {
			cName = t.Name()
		}
		rcName := []rune(cName)
		rcName[0] = unicode.ToLower(rcName[0])
		lcName = string(rcName)

		rmName := []rune(m.Name)
		rmName[0] = unicode.ToLower(rmName[0])
		lmName := string(rmName)


		r.GET( "/" + cName  + "/" + m.Name,callback)
		r.GET( "/" + lcName + "/" + lmName,callback)
		r.POST("/" + cName  + "/" + m.Name,callback)
		r.POST("/" + lcName + "/" + lmName,callback)
	}
}

