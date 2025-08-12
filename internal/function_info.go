package internal

import "reflect"

type functionInfo struct {
	inputTypes []reflect.Type
	outputType reflect.Type
}
