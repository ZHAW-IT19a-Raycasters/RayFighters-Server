package log

import "testing"

func TestPrependlevel(t *testing.T) {
	org := []interface{}{"1", "2", "3"}
	expected := []interface{}{Debug.String(), "1", "2", "3"}
	res := prependLevel(Debug, org)
	if len(res) != len(expected) {
		t.Error("expected length not reached")
	}
	for i := range res {
		if res[i] != expected[i] {
			t.Errorf("[%s] is not equals to [%s]", res[i], expected[i])
		}
	}
}
