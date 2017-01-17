package proxy

import (
	"fmt"
	"testing"
)

func TestStr2NonEmptySlice(t *testing.T) {
	testcases := []struct {
		Str string
		Res []string
	}{
		{
			"a,b,c",
			[]string{"a", "b", "c"},
		},
		{
			"  a, b , c ",
			[]string{"a", "b", "c"},
		},
		{
			"  a, b ,     , c ",
			[]string{"a", "b", "c"},
		},
		{
			"  a   ,  ,     ,   ,    b ,     , c ,   ",
			[]string{"a", "b", "c"},
		},
		{
			"          ,    a   ,  ,     ,   ,    b ,     , c ,   ",
			[]string{"a", "b", "c"},
		},
	}

	for i, c := range testcases {
		got := str2NonEmptySlice(c.Str, ",")
		gotStr := fmt.Sprintf("%v", got)
		expectedStr := fmt.Sprintf("%v", c.Res)

		if gotStr != expectedStr {
			t.Errorf("#%d expected %q, got %q", i, expectedStr, gotStr)
		}
	}
}
