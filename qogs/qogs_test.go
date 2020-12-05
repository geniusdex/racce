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
	assert.Equal(t, dataOnInt, SortOn(data, ".Integer"))
	assert.Equal(t, dataOnString, SortOn(data, ".String"))
}

func TestFilterEq(t *testing.T) {
	data := []testData{
		testData{5, "c"},
		testData{3, "b"},
		testData{4, "a"},
		testData{4, "d"},
	}
	dataFiltered := []interface{}{
		testData{4, "a"},
		testData{4, "d"},
	}
	assert.Equal(t, dataFiltered, FilterEq(data, ".Integer", "4"))
}

func TestValues(t *testing.T) {
	mapData := make(map[string]int)
	mapData["foo"] = 42
	mapData["bar"] = 27
	assert.ElementsMatch(t, []int{42, 27}, Values(mapData))

	array := []string{"test", "1", "2", "3"}
	assert.Equal(t, []interface{}{"test", "1", "2", "3"}, Values(array))
}

func TestReverse(t *testing.T) {
	assert.Equal(t, []interface{}{3, 2, 1}, Reverse([]int{1, 2, 3}))
	assert.Equal(t, []interface{}{"bar", "foo"}, Reverse([]string{"foo", "bar"}))
}

func TestContains(t *testing.T) {
	assert.True(t, Contains([]int{12, 72, 42}, 72))
	assert.False(t, Contains([]int{12, 72, 42}, 5))
	assert.True(t, Contains([]string{"foo", "bar"}, "foo"))
	assert.False(t, Contains([]string{"foo", "bar"}, "test"))
}

func TestLimit(t *testing.T) {
	assert.Equal(t, []interface{}{1, 2, 3}, Limit([]int{1, 2, 3, 4, 5}, 3))
	assert.Equal(t, []interface{}{1, 2, 3}, Limit([]int{1, 2, 3}, 5))
	assert.Equal(t, []interface{}{1, 2, 3}, Limit([]int{1, 2, 3}, 3))
}
