package utils

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArrayInsert(t *testing.T) {
	assert := assert.New(t)

	input := []string{"1", "3", "4"}

	result := StringArrayInsert(input, 1, "2")
	wanted := []string{"1", "2", "3", "4"}
	assert.True(reflect.DeepEqual(result, wanted))
}

func TestStringArrayInsert2(t *testing.T) {
	assert := assert.New(t)

	input := []string{"1", "2", "3", "5", "6"}

	result := StringArrayInsert(input, 3, "4")
	wanted := []string{"1", "2", "3", "4", "5", "6"}
	assert.True(reflect.DeepEqual(result, wanted))
}

func TestStringArrayReplace(t *testing.T) {
	assert := assert.New(t)

	input := []string{"1", "2", "3", "99", "5", "6"}

	result := StringArrayReplace(input, 3, "4")
	print(strings.Join(result, ",") + "\n")
	wanted := []string{"1", "2", "3", "4", "5", "6"}
	assert.True(reflect.DeepEqual(result, wanted))
}

func TestStringArrayReplace2(t *testing.T) {
	assert := assert.New(t)

	input := []string{"1", "2", "3", "4", "5", "99"}

	result := StringArrayReplace(input, 5, "6")
	print(strings.Join(result, ",") + "\n")
	wanted := []string{"1", "2", "3", "4", "5", "6"}
	assert.True(reflect.DeepEqual(result, wanted))
}
