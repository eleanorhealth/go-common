package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/eleanorhealth/go-common/v2/pkg/errs"
)

var env string

// initialized is a witness type returned by Setenv. Passing it to IsLocal,
// IsQA, or IsProd proves at compile time that Setenv was called first.
type initialized struct{}

func parse(e string) {
	switch e {
	case EnvLocal, EnvQA, EnvProd:
		return
	default:
		panic(fmt.Sprintf("invalid env %q", e))
	}
}

func Setenv(e string) initialized {
	parse(e)

	env = e

	return initialized{}
}

const (
	EnvLocal string = "local"
	EnvQA    string = "qa"
	EnvProd  string = "prod"
)

func Get[T bool | []byte | int | string](key string, defaultVal T) T {
	v, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}

	var ret T
	if err := parseInto(v, &ret); err != nil {
		return defaultVal
	}

	return ret
}

func parseInto(val string, ptr any) error {
	switch p := ptr.(type) {
	case *string:
		*p = val
	case *bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}

		*p = b
	case *int:
		n, err := strconv.Atoi(val)
		if err != nil {
			return err
		}

		*p = n
	case *[]byte:
		*p = []byte(val)
	default:
		return fmt.Errorf("unsupported type %T", ptr)
	}

	return nil
}

func GetExists[T bool | []byte | int | string](key string) (T, bool) {
	var v T

	_, exists := os.LookupEnv(key)
	if !exists {
		return v, false
	}

	return Get(key, v), true
}

// Different from Get[string] in that it returns the default value if the
// environment variable exists but is empty.
func GetString(key, defaultVal string) string {
	val := Get(key, defaultVal)
	if val == "" {
		return defaultVal
	}

	return val
}

func ParseStruct(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("env: ParseStruct requires a non-nil pointer to a struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	tagged := 0

	for i := range rt.NumField() {
		field := rt.Field(i)
		if !field.IsExported() || field.Type.Kind() == reflect.Struct {
			continue
		}

		key, ok := field.Tag.Lookup("env")
		if !ok {
			continue
		}

		tagged++

		val, exists := os.LookupEnv(key)
		if !exists {
			return fmt.Errorf("env: %q is not set", key)
		}

		if val == "" {
			continue
		}

		if err := parseInto(val, rv.Field(i).Addr().Interface()); err != nil {
			return errs.Wrapf(err, "env: parsing %q for field %s", key, field.Name)
		}
	}

	if tagged == 0 {
		return fmt.Errorf("env: ParseStruct found no env tags")
	}

	return nil
}

func IsLocal(_ initialized) bool {
	return env == EnvLocal
}

func IsQA(_ initialized) bool {
	return env == EnvQA
}

func IsProd(_ initialized) bool {
	return env == EnvProd
}
