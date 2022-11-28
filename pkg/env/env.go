package env

import (
	"os"
	"strconv"
)

const (
	EnvLocal string = "local"
	EnvQA    string = "qa"
	EnvProd  string = "prod"
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
	return Get("ENV", "") == EnvLocal
}

func IsQA() bool {
	return Get("ENV", "") == EnvQA
}

func IsProd() bool {
	return Get("ENV", "") == EnvProd

}
