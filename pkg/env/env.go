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

func Get[T bool | []byte | int | string](key string, d T) T {
	v, exists := os.LookupEnv(key)
	if !exists {
		return d
	}

	var ret T
	switch ptr := any(&ret).(type) {
	case *bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return d
		}

		*ptr = b

	case *[]byte:
		*ptr = []byte(v)

	case *int:
		i, err := strconv.Atoi(v)
		if err != nil {
			return d
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
