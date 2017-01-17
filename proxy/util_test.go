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
			"          ,    a\n   ,  ,   \n  ,   ,    b ,     , c ,   ",
			[]string{"a", "b", "c"},
		},
		{
			`
			:8001,
        	:8002,
			`,
			[]string{":8001", ":8002"},
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
