package interpreter

import (
	"github.com/mburakmmm/sky-lang/internal/runtime/skylib"
)

// addNativeStdlib adds native Go stdlib functions to environment
func addNativeStdlib(env *Environment) {
	// FS Module
	env.Set("fs_exists", createNativeFunc("fs_exists", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if path, ok := args[0].(*String); ok {
				exists := skylib.FSExists(path.Value)
				return &Boolean{Value: exists}, nil
			}
		}
		return &Boolean{Value: false}, nil
	}))

	env.Set("fs_read_text", createNativeFunc("fs_read_text", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if path, ok := args[0].(*String); ok {
				content, err := skylib.FSReadText(path.Value)
				if err != nil {
					return &Nil{}, &RuntimeError{Message: err.Error()}
				}
				return &String{Value: content}, nil
			}
		}
		return &String{Value: ""}, nil
	}))

	env.Set("fs_write_text", createNativeFunc("fs_write_text", func(args []Value) (Value, error) {
		if len(args) >= 2 {
			if path, ok := args[0].(*String); ok {
				if data, ok := args[1].(*String); ok {
					err := skylib.FSWriteText(path.Value, data.Value)
					if err != nil {
						return &Boolean{Value: false}, &RuntimeError{Message: err.Error()}
					}
					return &Boolean{Value: true}, nil
				}
			}
		}
		return &Boolean{Value: false}, nil
	}))

	// OS Module
	env.Set("os_getenv", createNativeFunc("os_getenv", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if key, ok := args[0].(*String); ok {
				value := skylib.OSGetEnv(key.Value)
				return &String{Value: value}, nil
			}
		}
		return &String{Value: ""}, nil
	}))

	env.Set("os_getcwd", createNativeFunc("os_getcwd", func(args []Value) (Value, error) {
		cwd, err := skylib.OSGetcwd()
		if err != nil {
			return &String{Value: ""}, &RuntimeError{Message: err.Error()}
		}
		return &String{Value: cwd}, nil
	}))

	env.Set("os_platform", createNativeFunc("os_platform", func(args []Value) (Value, error) {
		return &String{Value: skylib.OSPlatform()}, nil
	}))

	// Crypto Module
	env.Set("crypto_sha256", createNativeFunc("crypto_sha256", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if data, ok := args[0].(*String); ok {
				hash := skylib.CryptoSHA256([]byte(data.Value))
				return &String{Value: hash}, nil
			}
		}
		return &String{Value: ""}, nil
	}))

	env.Set("crypto_md5", createNativeFunc("crypto_md5", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if data, ok := args[0].(*String); ok {
				hash := skylib.CryptoMD5([]byte(data.Value))
				return &String{Value: hash}, nil
			}
		}
		return &String{Value: ""}, nil
	}))

	// JSON Module
	env.Set("json_encode", createNativeFunc("json_encode", func(args []Value) (Value, error) {
		if len(args) > 0 {
			jsonStr, err := skylib.JSONEncode(convertToGo(args[0]))
			if err != nil {
				return &String{Value: ""}, &RuntimeError{Message: err.Error()}
			}
			return &String{Value: jsonStr}, nil
		}
		return &String{Value: ""}, nil
	}))

	env.Set("json_decode", createNativeFunc("json_decode", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if jsonStr, ok := args[0].(*String); ok {
				obj, err := skylib.JSONDecode(jsonStr.Value)
				if err != nil {
					return &Nil{}, &RuntimeError{Message: err.Error()}
				}
				return convertFromGo(obj), nil
			}
		}
		return &Nil{}, nil
	}))

	// Time Module
	env.Set("time_now", createNativeFunc("time_now", func(args []Value) (Value, error) {
		ts := skylib.TimeNow()
		return &Integer{Value: ts}, nil
	}))

	env.Set("time_sleep", createNativeFunc("time_sleep", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if ms, ok := args[0].(*Integer); ok {
				skylib.TimeSleep(int(ms.Value))
			}
		}
		return &Nil{}, nil
	}))

	// Rand Module
	env.Set("rand_int", createNativeFunc("rand_int", func(args []Value) (Value, error) {
		if len(args) > 0 {
			if max, ok := args[0].(*Integer); ok {
				val := skylib.RandIntN(int(max.Value))
				return &Integer{Value: int64(val)}, nil
			}
		}
		return &Integer{Value: 0}, nil
	}))

	env.Set("rand_uuid", createNativeFunc("rand_uuid", func(args []Value) (Value, error) {
		uuid := skylib.RandUUID()
		return &String{Value: uuid}, nil
	}))
}

// createNativeFunc creates a native function wrapper
func createNativeFunc(name string, fn func([]Value) (Value, error)) *Function {
	return &Function{
		Name: name,
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok {
				return fn(list.Elements)
			}
			return fn([]Value{})
		},
	}
}

// convertToGo converts Sky value to Go interface{}
func convertToGo(val Value) interface{} {
	switch v := val.(type) {
	case *Integer:
		return v.Value
	case *Float:
		return v.Value
	case *String:
		return v.Value
	case *Boolean:
		return v.Value
	case *List:
		arr := make([]interface{}, len(v.Elements))
		for i, elem := range v.Elements {
			arr[i] = convertToGo(elem)
		}
		return arr
	case *Dict:
		m := make(map[string]interface{})
		for k, val := range v.Pairs {
			m[k] = convertToGo(val)
		}
		return m
	case *Nil:
		return nil
	default:
		return nil
	}
}

// convertFromGo converts Go interface{} to Sky value
func convertFromGo(obj interface{}) Value {
	if obj == nil {
		return &Nil{}
	}

	switch v := obj.(type) {
	case int:
		return &Integer{Value: int64(v)}
	case int64:
		return &Integer{Value: v}
	case float64:
		return &Float{Value: v}
	case string:
		return &String{Value: v}
	case bool:
		return &Boolean{Value: v}
	case []interface{}:
		elements := make([]Value, len(v))
		for i, elem := range v {
			elements[i] = convertFromGo(elem)
		}
		return &List{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[string]Value)
		for k, val := range v {
			pairs[k] = convertFromGo(val)
		}
		return &Dict{Pairs: pairs}
	default:
		return &Nil{}
	}
}
