package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("FOO", "1")
	assert.Equal("1", Get("FOO", ""))
	assert.Equal([]byte("1"), Get("FOO", []byte("")))
	assert.Equal(1, Get("FOO", 0))
	assert.Equal(true, Get("FOO", false))
}

func TestGet_default(t *testing.T) {
	assert := assert.New(t)

	os.Clearenv()
	assert.Equal("baz", Get("FOO", "baz"))
	assert.Equal([]byte("baz"), Get("FOO", []byte("baz")))
	assert.Equal(true, Get("FOO", true))
}

func TestGetExists(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("FOO", "1")

	v, exists := GetExists[string]("FOO")
	assert.Equal("1", v)
	assert.True(exists)

	v, exists = GetExists[string]("BAR")
	assert.Equal("", v)
	assert.False(exists)
}

func TestIsLocal(t *testing.T) {
	assert := assert.New(t)

	Setenv("local")
	assert.True(IsLocal())
	assert.False(IsQA())
	assert.False(IsProd())
}

func TestIsQA(t *testing.T) {
	assert := assert.New(t)

	Setenv("qa")
	assert.False(IsLocal())
	assert.True(IsQA())
	assert.False(IsProd())
}

func TestIsProd(t *testing.T) {
	assert := assert.New(t)

	Setenv("prod")
	assert.False(IsLocal())
	assert.False(IsQA())
	assert.True(IsProd())
}

func TestIsX_panic(t *testing.T) {
	assert := assert.New(t)

	env = ""

	assert.PanicsWithValue("invalid env \"\"", func() {
		IsLocal()
	})
	assert.PanicsWithValue("invalid env \"\"", func() {
		IsQA()
	})
	assert.PanicsWithValue("invalid env \"\"", func() {
		IsProd()
	})
}

func TestSetenv_panic(t *testing.T) {
	assert := assert.New(t)

	assert.PanicsWithValue("invalid env \"unknown\"", func() {
		Setenv("unknown")
	})
}
