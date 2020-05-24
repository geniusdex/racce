package qogs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPathEntry struct {
	Name string
}

type testPathStruct struct {
	ID      int
	Names   []string
	Entry   *testPathEntry
	Entries map[string]*testPathEntry
}

func TestPath(t *testing.T) {
	assert := assert.New(t)

	data := &testPathStruct{
		5,
		[]string{"foo", "bar"},
		&testPathEntry{"Entry"},
		map[string]*testPathEntry{
			"foo": &testPathEntry{"foo"},
			"bar": &testPathEntry{"bar"},
		},
	}

	assert.Equal(5, Path(data, ".ID"))
	assert.Equal("Entry", Path(data, ".Entry.Name"))

	assert.Equal(2, Path(data, "len .Names"))

	assert.Equal("entry", Path(data, "tolower .Entry.Name"))
}
