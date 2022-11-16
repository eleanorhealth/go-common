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

func TestIsLocal(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("ENV", "local")
	assert.True(IsLocal())
	assert.False(IsQA())
	assert.False(IsProd())
}

func TestIsQA(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("ENV", "qa")
	assert.False(IsLocal())
	assert.True(IsQA())
	assert.False(IsProd())
}

func TestIsProd(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("ENV", "prod")
	assert.False(IsLocal())
	assert.False(IsQA())
	assert.True(IsProd())
}
