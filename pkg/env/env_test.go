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

func TestParseStruct(t *testing.T) {
	assert := assert.New(t)

	type nested struct{ X string }
	type cfg struct {
		Str    string `env:"PS_STR"`
		Bool   bool   `env:"PS_BOOL"`
		Int    int    `env:"PS_INT"`
		Bytes  []byte `env:"PS_BYTES"`
		NoTag  string
		Empty  string `env:"PS_EMPTY"`
		nested        // embedded struct: skipped
		//nolint:unused
		unexported string `env:"PS_UNEXPORTED"`
	}

	t.Setenv("PS_STR", "hello")
	t.Setenv("PS_BOOL", "true")
	t.Setenv("PS_INT", "42")
	t.Setenv("PS_BYTES", "data")
	t.Setenv("PS_EMPTY", "")

	var c cfg
	err := ParseStruct(&c)
	assert.NoError(err)
	assert.Equal("hello", c.Str)
	assert.Equal(true, c.Bool)
	assert.Equal(42, c.Int)
	assert.Equal([]byte("data"), c.Bytes)
	assert.Equal("", c.NoTag)
	assert.Equal("", c.Empty)
}

func TestParseStruct_envDefault(t *testing.T) {
	assert := assert.New(t)

	type cfg struct {
		Unset   string `env:"PS_DEF_UNSET" envDefault:"fallback"`
		Empty   string `env:"PS_DEF_EMPTY" envDefault:"fallback2"`
		Present string `env:"PS_DEF_PRESENT" envDefault:"ignored"`
	}

	os.Unsetenv("PS_DEF_UNSET")
	t.Setenv("PS_DEF_EMPTY", "")
	t.Setenv("PS_DEF_PRESENT", "actual")

	var c cfg
	err := ParseStruct(&c)
	assert.NoError(err)
	assert.Equal("fallback", c.Unset)
	assert.Equal("fallback2", c.Empty)
	assert.Equal("actual", c.Present)
}

func TestParseStruct_noTags(t *testing.T) {
	assert := assert.New(t)

	type cfg struct{ Foo string }

	err := ParseStruct(&cfg{})
	assert.EqualError(err, "env: ParseStruct found no env tags")
}

func TestParseStruct_unset(t *testing.T) {
	assert := assert.New(t)

	type cfg struct {
		Foo string `env:"PS_UNSET_FOO"`
	}

	os.Unsetenv("PS_UNSET_FOO")

	err := ParseStruct(&cfg{})
	assert.EqualError(err, `env: "PS_UNSET_FOO" is not set`)
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
