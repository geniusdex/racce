package qogs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	Integer int
	String  string
}

func TestElem(t *testing.T) {
	root := testData{}
	if _, ok := Elem(root).(testData); !ok {
		t.Errorf("Elem() does not return struct verbatim")
	}
	if _, ok := Elem(&root).(testData); !ok {
		t.Errorf("Elem() does not unpack struct pointer")
	}
	if _, ok := Elem(root.Integer).(int); !ok {
		t.Errorf("Elem() does not return integer verbatim")
	}
	if _, ok := Elem(&root.String).(string); !ok {
		t.Errorf("Elem() does not unpack string pointer")
	}
}

func TestKeys(t *testing.T) {
	data := make(map[string]int)
	data["foo"] = 42
	data["bar"] = 27
	assert.ElementsMatch(t, []string{"foo", "bar"}, Keys(data))
}

func TestPath(t *testing.T) {
	assert.Equal(t, 2, Path(testData{2, ""}, "Integer"))
	assert.Equal(t, "test", Path(testData{2, "test"}, "String"))
}

func TestSortOn(t *testing.T) {
	data := []testData{
		testData{5, "c"},
		testData{3, "b"},
		testData{4, "a"},
	}
	dataOnInt := []interface{}{
		testData{3, "b"},
		testData{4, "a"},
		testData{5, "c"},
	}
	dataOnString := []interface{}{
		testData{4, "a"},
		testData{3, "b"},
		testData{5, "c"},
	}
	assert.Equal(t, dataOnInt, SortOn(data, "Integer"))
	assert.Equal(t, dataOnString, SortOn(data, "String"))
}

func TestValues(t *testing.T) {
	mapData := make(map[string]int)
	mapData["foo"] = 42
	mapData["bar"] = 27
	assert.ElementsMatch(t, []int{42, 27}, Values(mapData))

	array := []string{"test", "1", "2", "3"}
	assert.Equal(t, []interface{}{"test", "1", "2", "3"}, Values(array))
}
