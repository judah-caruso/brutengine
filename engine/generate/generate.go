package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/judah-caruso/brutengine/engine"
)

// Interfaces to generate wasm wrappers/exposers
var interfaces = []interface{}{
	(*engine.IConfig)(nil),
	(*engine.IPlatform)(nil),
	(*engine.IInput)(nil),
	(*engine.IGraphics)(nil),
	(*engine.IAsset)(nil),
}

var apiVersion = "0.0.1"

var fileSkeleton = `package engine

import (
	"context"
	"github.com/tetratelabs/wazero/api"
)

{{ .Exposer }}

// Wasm wrappers for {{ .Namespace }}
{{ .Wrappers }}`

var exposerSkeleton = `
func (a *{{ .Namespace }}) Expose(wasm *WasmRuntime) {
{{ range $_, $name := .Functions -}}
	wasm.ConvertAndExpose("{{ $.Namespace }}{{ $name }}", a.{{ $name }}, wasm{{ $name }})
{{ end }}
}`

var wrapperSkeleton = `
// Calls {{ .Namespace }}.{{ .GoName }}
func {{ .WasmName }}(ctx context.Context, m api.Module, stack []WasmValue) {
{{ .WasmArguments -}}
{{ .GoCall -}}
{{ .WasmReturns -}}
}`

type (
	ApiType string
	Api     struct {
		Version string                `json:"version"`
		Structs map[ApiType][]ApiType `json:"structs"`
		Exports []Export              `json:"exports"`
	}
	Export struct {
		Namespace string     `json:"namespace"`
		Functions []Function `json:"functions"`
	}
	Function struct {
		Name string    `json:"name"`
		Args []ApiType `json:"args"`
		Rets []ApiType `json:"rets"`
	}
)

const (
	ApiBool  ApiType = "bool"
	ApiUint  ApiType = "u32"
	ApiInt   ApiType = "i32"
	ApiFloat ApiType = "f32"
)

var structTypes = map[ApiType][]ApiType{
	"string": {ApiUint, ApiUint},
}

func goTypeToApi(t reflect.Type) ApiType {
	k := t.Kind()
	switch k {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return ApiInt
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return ApiUint
	case reflect.Bool:
		return ApiBool
	case reflect.Float32, reflect.Float64:
		return ApiFloat
	case reflect.String:
		return ApiType("string")
	case reflect.Struct:
		name := ApiType(t.Name())
		_, ok := structTypes[name]
		if ok {
			return name
		}

		comp := make([]ApiType, 0)
		for i := 0; i < t.NumField(); i += 1 {
			f := t.Field(i)
			comp = append(comp, goTypeToApi(f.Type))
		}

		structTypes[name] = comp
		return name
	default:
		panic(fmt.Sprintf("Go kind %s cannot be converted to an api type", t))
	}
}

func decodeFromStack(t ApiType) string {
	switch t {
	case ApiInt:
		return "api.DecodeI32"
	case ApiBool, ApiUint:
		return "api.DecodeU32"
	case ApiFloat:
		return "api.DecodeF32"
	default:
		panic("unreachable")
	}
}

func encodeToStack(t ApiType) string {
	switch t {
	case ApiInt:
		return "api.EncodeI32"
	case ApiBool, ApiUint:
		return "api.EncodeU32"
	case ApiFloat:
		return "api.EncodeF32"
	default:
		panic("unreachable")
	}
}

func normalizeType(t ApiType) string {
	switch t {
	case ApiInt:
		return "int32"
	case ApiUint:
		return "uint32"
	case ApiBool:
		return "boolToU32"
	case ApiFloat:
		return "float32"
	default:
		panic("unreachable")
	}
}

func wasmToGo(t ApiType, variable string) string {
	switch t {
	case ApiInt:
		return fmt.Sprintf("int32(%s)", variable)
	case ApiUint:
		return fmt.Sprintf("uint32(%s)", variable)
	case ApiBool:
		return fmt.Sprintf("u32ToBool(%s)", variable)
	case ApiFloat:
		return fmt.Sprintf("float32(%s)", variable)
	case "string":
		return fmt.Sprintf("readWasmString(m.Memory(), %[1]s_0, %[1]s_1)", variable)
	default:
		comp, ok := structTypes[t]
		if !ok {
			panic("attempt to convert non-existant type:" + t)
		}

		buf := bytes.Buffer{}
		buf.WriteString(string(t))
		buf.WriteByte('{')

		for i, c := range comp {
			buf.WriteString(wasmToGo(c, fmt.Sprintf("%s_%d", variable, i)))
			if i < len(comp)-1 {
				buf.WriteString(", ")
			}
		}

		buf.WriteByte('}')
		return buf.String()
	}
}

func generateJsonApi(types []reflect.Type) (*Api, error) {
	jsonApi := Api{
		Version: apiVersion,
	}

	for _, typ := range types {
		funcs := make([]Function, 0)
		for mi := 0; mi < typ.NumMethod(); mi += 1 {
			var (
				m    = typ.Method(mi)
				name = m.Name
				args = make([]ApiType, 0)
				rets = make([]ApiType, 0)
			)

			for ai := 0; ai < m.Type.NumIn(); ai += 1 {
				arg := m.Type.In(ai)
				argType := goTypeToApi(arg)
				args = append(args, argType)
			}

			for ri := 0; ri < m.Type.NumOut(); ri += 1 {
				ret := m.Type.Out(ri)
				rets = append(rets, goTypeToApi(ret))
			}

			funcs = append(funcs, Function{
				Name: name,
				Args: args,
				Rets: rets,
			})
		}

		namespace := typ.Name()[1:]
		jsonApi.Exports = append(jsonApi.Exports, Export{
			Namespace: namespace,
			Functions: funcs,
		})
	}

	jsonApi.Structs = structTypes
	return &jsonApi, nil
}

func main() {
	itypes := make([]reflect.Type, 0)
	for _, iface := range interfaces {
		itypes = append(itypes, reflect.TypeOf(iface).Elem())
	}

	var (
		api *Api
		err error
	)

	fmt.Println("generating json api")

	{
		api, err = generateJsonApi(itypes)
		if err != nil {
			panic(err)
		}

		out, err := json.MarshalIndent(api, "", "  ")
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("../engine_api.json", out, fs.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	fileTemplate, err := template.New("file").Parse(fileSkeleton)
	if err != nil {
		panic(err)
	}

	wrapperTemplate, err := template.New("wrapper").Parse(wrapperSkeleton)
	if err != nil {
		panic(err)
	}

	exposerTemplate, err := template.New("exposer").Parse(exposerSkeleton)
	if err != nil {
		panic(err)
	}

	fmt.Println("generating wrappers")

	{
		for _, export := range api.Exports {
			wrapperBuf := bytes.Buffer{}

			// generate wrapper functions
			for _, fn := range export.Functions {
				var data struct {
					Namespace     string
					GoName        string
					GoCall        string
					WasmName      string
					WasmArguments string
					WasmReturns   string
				}

				data.Namespace = export.Namespace
				data.GoName = fn.Name
				data.WasmName = "wasm" + fn.Name

				{ // generate pulling wasm arguments from the stack
					argBuf := bytes.Buffer{}
					stackIdx := 0
					for i, arg := range fn.Args {
						name := fmt.Sprintf("arg%d", i)
						comp, isStruct := structTypes[arg]
						if isStruct {
							for ci, t := range comp {
								fmt.Fprintf(&argBuf, "\t%s_%d := %s(stack[%d])\n", name, ci, decodeFromStack(t), stackIdx)
								stackIdx += 1
							}
						} else {
							fmt.Fprintf(&argBuf, "\t%s := %s(stack[%d])\n", name, decodeFromStack(arg), stackIdx)
							stackIdx += 1
						}
					}

					data.WasmArguments = argBuf.String()
				}

				{ // generate wasm to go conversions and calling into engine code
					callBuf := bytes.Buffer{}
					callBuf.WriteByte('\t')

					if len(fn.Rets) > 0 {
						for i := range fn.Rets {
							fmt.Fprintf(&callBuf, "r%d", i)
							if i < len(fn.Rets)-1 {
								callBuf.WriteString(", ")
							}
						}

						callBuf.WriteString(" := ")
					}

					fmt.Fprintf(&callBuf, "brut.%s.%s(", export.Namespace, fn.Name)
					if len(fn.Args) > 0 {
						callBuf.WriteByte('\n')
					}

					for i, arg := range fn.Args {
						callBuf.WriteString("\t\t")
						callBuf.WriteString(wasmToGo(arg, fmt.Sprintf("arg%d", i)))
						callBuf.WriteString(",\n")
					}

					if len(fn.Args) > 0 {
						callBuf.WriteByte('\t')
					}

					callBuf.WriteString(")\n")

					data.GoCall = callBuf.String()
				}

				{ // generate pushing return values to wasm stack
					retBuf := bytes.Buffer{}
					stackIdx := 0

					for i, ret := range fn.Rets {
						comp, isStruct := structTypes[ret]
						if isStruct {
							for ci, t := range comp {
								fmt.Fprintf(&retBuf, "\tstack[%d] = %s(%s(r%d))\n", stackIdx, encodeToStack(t), normalizeType(t), ci+i)
								stackIdx += 1
							}
						} else {
							fmt.Fprintf(&retBuf, "\tstack[%d] = %s(%s(r%d))\n", stackIdx, encodeToStack(ret), normalizeType(ret), i)
							stackIdx += 1
						}
					}

					data.WasmReturns = retBuf.String()
				}

				err := wrapperTemplate.Execute(&wrapperBuf, data)
				if err != nil {
					panic(err)
				}
			}

			var exposerBuf bytes.Buffer
			var exposerData struct {
				Namespace string
				Functions []string
			}

			exposerData.Namespace = export.Namespace
			for _, fn := range export.Functions {
				exposerData.Functions = append(exposerData.Functions, fn.Name)
			}

			err := exposerTemplate.Execute(&exposerBuf, exposerData)
			if err != nil {
				panic(err)
			}

			var templateData struct {
				Namespace string
				Exposer   string
				Wrappers  string
			}

			templateData.Namespace = export.Namespace
			templateData.Exposer = exposerBuf.String()
			templateData.Wrappers = wrapperBuf.String()

			var file bytes.Buffer

			err = fileTemplate.Execute(&file, templateData)
			if err != nil {
				panic(err)
			}

			filename := "wasm_" + strings.ToLower(export.Namespace) + ".gen.go"

			formatted, err := format.Source(file.Bytes())
			if err != nil {
				panic(err)
			}

			formatted = append([]byte("// Code generated by 'go generate ./...'; DO NOT EDIT.\n"), formatted...)

			err = os.WriteFile(filename, formatted, os.ModePerm)
			if err != nil {
				panic(err)
			}

			fmt.Println("\tgenerated", filename)
		}
	}
}

func generateWrappers(ns string, goType, wasmType reflect.Type) (string, error) {
	for i := 0; i < goType.NumMethod(); i += 1 {
		goMethod := goType.Method(i)
		wasmMethod, ok := wasmType.MethodByName(goMethod.Name)
		if !ok {
			return "", fmt.Errorf("%s does not match %s! Function %q is missing from the interface", wasmType.Name(), goType.Name(), goMethod.Name)
		}

		fmt.Println(wasmMethod.Name)
	}
	return "", nil
}
