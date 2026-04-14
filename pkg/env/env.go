package env

import (
	"fmt"
	"os"
	"strconv"
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
	switch ptr := any(&ret).(type) {
	case *bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return defaultVal
		}

		*ptr = b

	case *[]byte:
		*ptr = []byte(v)

	case *int:
		i, err := strconv.Atoi(v)
		if err != nil {
			return defaultVal
		}

		*ptr = i

	case *string:
		*ptr = v
	}

	return ret
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

func IsLocal(_ initialized) bool {
	return env == EnvLocal
}

func IsQA(_ initialized) bool {
	return env == EnvQA
}

func IsProd(_ initialized) bool {
	return env == EnvProd
}
