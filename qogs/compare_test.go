package qogs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompareBoolean(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, Compare(false, false))
	assert.Less(Compare(false, true), 0)
	assert.Greater(Compare(true, false), 0)
	assert.Equal(0, Compare(true, true))
}

func TestCompareNumeric(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, Compare(5, 5))
	assert.Less(Compare(13, 37), 0)
	assert.Greater(Compare(42, 27), 0)

	var smallInt int8 = -12
	var largeInt int32 = 12345
	var smallUint uint = 17
	var largeUint uint64 = 1234567890
	var smallFloat float32 = -1.2e-7
	var largeFloat float64 = 3.14159e112

	assert.Less(Compare(smallInt, largeInt), 0)
	assert.Less(Compare(smallUint, &largeUint), 0)
	assert.Less(Compare(&smallFloat, largeFloat), 0)
	assert.Less(Compare(&smallInt, &smallUint), 0)
	assert.Less(Compare(largeInt, largeUint), 0)
	assert.Less(Compare(smallInt, smallFloat), 0)
	assert.Less(Compare(smallFloat, smallUint), 0)
	assert.Less(Compare(largeInt, largeFloat), 0)
	assert.Less(Compare(largeUint, largeFloat), 0)
}

func TestCompareStrings(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, Compare("foo", "foo"))
	assert.Less(Compare("foo", "food"), 0)
	assert.Greater(Compare("foo", "bar"), 0)
}

func TestCompareTime(t *testing.T) {
	assert := assert.New(t)

	history := time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)
	future := time.Date(2040, 7, 8, 9, 10, 11, 12, time.UTC)

	assert.Equal(Compare(history, &history), 0)
	assert.Less(Compare(history, future), 0)
	assert.Greater(Compare(future, history), 0)
}

func TestCompareStringConversions(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, Compare("5", 5))
	assert.Equal(0, Compare(5, "5"))
	assert.NotEqual(0, Compare("five", 5))
	assert.NotEqual(0, Compare(5, "five"))
}

type testCompareStructPlainReceiver struct {
	Integer int
}

func (me testCompareStructPlainReceiver) CompareTo(other testCompareStructPlainReceiver) int {
	return me.Integer - other.Integer
}

type testCompareStructPtrReceiver struct {
	Integer int
}

func (me *testCompareStructPtrReceiver) CompareTo(other *testCompareStructPtrReceiver) int {
	return me.Integer - other.Integer
}

func TestCompareStruct(t *testing.T) {
	assert := assert.New(t)

	smallPlain := testCompareStructPlainReceiver{13}
	largePlain := testCompareStructPlainReceiver{37}

	assert.Equal(0, Compare(smallPlain, smallPlain))
	assert.Less(Compare(smallPlain, &largePlain), 0)
	assert.Greater(Compare(&largePlain, smallPlain), 0)
	assert.Equal(0, Compare(&largePlain, &largePlain))

	smallPtr := testCompareStructPtrReceiver{13}
	largePtr := testCompareStructPtrReceiver{37}

	assert.Equal(0, Compare(&smallPtr, &smallPtr))
	assert.Less(Compare(&smallPtr, &largePtr), 0)
	assert.Greater(Compare(&largePtr, &smallPtr), 0)
	assert.Equal(0, Compare(&largePtr, &largePtr))
}

type testCompareIncomparableStruct struct{}

func TestCompareIncompatible(t *testing.T) {
	assert := assert.New(t)
	assert.Panics(func() { Compare(5, true) })
	assert.Panics(func() { Compare(false, 2) })
	assert.Panics(func() { Compare(0, []int{}) })
	assert.Panics(func() { Compare(make(map[int]string), false) })
	assert.Panics(func() { Compare(testCompareIncomparableStruct{}, 2) })
}
