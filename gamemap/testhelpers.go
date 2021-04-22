package gamemap

import (
	"bytes"
	"strconv"
	"strings"
)

// intoMapString converts given 1d slice into map string, a padding will be included and each position of given slice is interpreted as a new row.
// Is a helper function used in testing.
func intoMapString(mapDim int, mapSlice []string) string {
	var map2d bytes.Buffer
	for i := 0; i < headerLength-1; i++ {
		map2d.WriteString("\n")
	}

	map2d.WriteString(strconv.Itoa(mapDim))
	map2d.WriteString("\n")
	dim := len(mapSlice)
	for i := 0; i < dim; i++ {
		map2d.WriteString(mapSlice[i])
		if i != dim-1 {
			map2d.WriteString("\n")
		}
	}
	return map2d.String()
}

// intoMapReader converts given slice into proper map format an return a *string.Reader on it.
func intoMapReader(mapDim int, mapSlice []string) *strings.Reader {
	return strings.NewReader(intoMapString(mapDim, mapSlice))
}
