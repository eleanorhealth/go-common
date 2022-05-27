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

func Get[T bool | []byte | string](key string, d T) T {
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

	case *string:
		*ptr = v
	}

	return ret
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
