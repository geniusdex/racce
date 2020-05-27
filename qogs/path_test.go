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

	// Selectors
	assert.Equal(5, Path(data, ".ID"))
	assert.Equal("Entry", Path(data, ".Entry.Name"))

	// Numeric constants
	assert.Equal(int64(27), Path(data, "27"))
	assert.InEpsilon(3.1415926535, Path(data, "3.1415926535"), 1e-9)
	assert.Equal(int64(-5), Path(data, "-5"))

	// Function 'len'
	assert.Equal(2, Path(data, "len .Names"))

	// Function 'tolower'
	assert.Equal("entry", Path(data, "tolower .Entry.Name"))
}
