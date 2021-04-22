package gamemap

import (
	"strings"
	"testing"
)

// Test if mapString produces expected result.
func TestIntoMapString(t *testing.T) {
	expected := strings.Repeat("\n", (headerLength-1)) + "2\n00\n0"
	actual := intoMapString(2, []string{"00", "0"})
	if expected != actual {
		t.Errorf("\nfalse result: (\nexpected:\n%s\nactual:\n%s", expected, actual)
	}
}
