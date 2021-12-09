package geerpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

//===============================服务注册========================================
/*
服务注册是以结构体为单位的。一个结构体对应一个service
service
|_methodA
|_methodB
|_...
*/
/*
	方法
*/
// methodType 包含了一个被远程调用的方法的完整信息
// todo: method,Method,func 这几个之间有什么区别
type methodType struct {
	method    reflect.Method // 被远程调用的方法本身
	ArgType   reflect.Type // "方法"的参数类型
	ReplyType reflect.Type // "方法"的返回类型
	numCalls  uint64 // "方法"的调用次数
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

// newArgv 新建一个argv的实例
func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	// arg may be a pointer type, or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

// newReplyv 新建一个replyv的实例
func (m *methodType) newReplyv() reflect.Value {
	// reply must be a pointer type
	replyv := reflect.New(m.ReplyType.Elem())//Elem(): 获取反射对应的
	switch m.ReplyType.Elem().Kind() {
	// 啥意思
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

/*
	service
*/

type service struct {
	name   string	// 映射的结构体的名称
	typ    reflect.Type //结构体的类型
	rcvr   reflect.Value //rcvr 即结构体的实例本身
	method map[string]*methodType// 存储映射的结构体的所有符合条件的方法
}

//newService 创建一个关于rcvr的service，一个rcvr对应一个service
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

// registerMethods 注册结构体的所有方法
// 两个导出或内置类型的入参（反射时为 3 个，第 0 个是自身，类似于 python 的 self，java 中的 this）
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}// 方法的输入参数必须为3，输出参数必须为1
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}// 方法的输出参数的类型必须是error
		argType, replyType := mType.In(1), mType.In(2)//获取方法的第一个和第2个参数
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// call 调用service的对应method
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

