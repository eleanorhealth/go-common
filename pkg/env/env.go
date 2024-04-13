package env

import (
	"fmt"
	"os"
	"strconv"
)

var env Env

type Env string

func (e Env) validate() {
	switch e {
	case EnvLocal, EnvQA, EnvProd:
		return
	default:
		panic(fmt.Sprintf("invalid env \"%s\"", e))
	}
}

func Setenv(e Env) {
	e.validate()

	env = e
}

const (
	EnvLocal Env = "local"
	EnvQA    Env = "qa"
	EnvProd  Env = "prod"
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
	val := Get[string](key, defaultVal)
	if val == "" {
		return defaultVal
	}

	return val
}

func IsLocal() bool {
	env.validate()

	return env == EnvLocal
}

func IsQA() bool {
	env.validate()

	return env == EnvQA
}

func IsProd() bool {
	env.validate()

	return env == EnvProd

}
