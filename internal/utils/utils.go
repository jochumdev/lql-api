package utils

import "os"

// BinarySearch search for a given byte in the bytearray
func BinarySearch(haystack []byte, needle byte) (result int) {
	result = -1

	var i = 0
	for _, b := range haystack {
		if b == needle {
			result = i
			break
		}
		i++
	}

	return result
}

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func StringArrayInsert(array []string, i int, insert string) []string {
	return append(array[:i], append([]string{insert}, array[i:]...)...)
}

func StringArrayReplace(array []string, i int, new string) []string {
	if len(array) == i {
		return append(array[i:1], new)
	}
	return append(array[:i], append([]string{new}, array[i+1:]...)...)
}
