package gocmd_test

import (
	"testing"

	"github.com/nikole-dunixi/add-direnv-gotool/internal/gocmd"
	"github.com/stretchr/testify/assert"
)

func TestModFilename(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		filename, err := gocmd.ModFilename("/foo/bar/baz", "newtool")
		assert.NoError(t, err)
		assert.Equal(t, "/foo/bar/baz/newtool.mod", filename)
	})
	t.Run("has extension", func(t *testing.T) {
		filename, err := gocmd.ModFilename("/foo/bar/baz", "newtool.mod")
		assert.NoError(t, err)
		assert.Equal(t, "/foo/bar/baz/newtool.mod", filename)
	})
	t.Run("has extension uppercase", func(t *testing.T) {
		filename, err := gocmd.ModFilename("/foo/bar/baz", "newtool.MOD")
		assert.NoError(t, err)
		assert.Equal(t, "/foo/bar/baz/newtool.MOD", filename)
	})
	t.Run("empty directory", func(t *testing.T) {
		filename, err := gocmd.ModFilename("", "newtool.mod")
		assert.NoError(t, err)
		assert.Equal(t, "newtool.mod", filename)
	})
	t.Run("empty filename fails", func(t *testing.T) {
		_, err := gocmd.ModFilename("/foo/bar/baz", "")
		assert.Error(t, err)
	})
}
