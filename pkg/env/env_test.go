package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	assert := assert.New(t)

	t.Setenv("FOO", "1")
	t.Setenv("BAR", "")
	assert.Equal("1", Get("FOO", ""))
	assert.Equal([]byte("1"), Get("FOO", []byte("")))
	assert.Equal(1, Get("FOO", 0))
	assert.Equal(true, Get("FOO", false))
	assert.Equal("", Get("BAR", "default"))
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

	t.Setenv("FOO", "1")

	v, exists := GetExists[string]("FOO")
	assert.Equal("1", v)
	assert.True(exists)

	v, exists = GetExists[string]("BAR")
	assert.Equal("", v)
	assert.False(exists)
}

func TestGetString(t *testing.T) {
	assert := assert.New(t)

	t.Setenv("FOO", "1")
	t.Setenv("BAR", "")

	v := GetString("FOO", "default1")
	assert.Equal("1", v)

	v = Get("FOO", "default1")
	assert.Equal("1", v)

	v = GetString("BAR", "default2")
	assert.Equal("default2", v)
}

func TestIsLocal(t *testing.T) {
	assert := assert.New(t)

	token := Setenv("local")
	assert.True(IsLocal(token))
	assert.False(IsQA(token))
	assert.False(IsProd(token))
}

func TestIsQA(t *testing.T) {
	assert := assert.New(t)

	token := Setenv("qa")
	assert.False(IsLocal(token))
	assert.True(IsQA(token))
	assert.False(IsProd(token))
}

func TestIsProd(t *testing.T) {
	assert := assert.New(t)

	token := Setenv("prod")
	assert.False(IsLocal(token))
	assert.False(IsQA(token))
	assert.True(IsProd(token))
}

func TestSetenv_panic(t *testing.T) {
	assert := assert.New(t)

	assert.PanicsWithValue("invalid env \"unknown\"", func() {
		Setenv("unknown")
	})
}
