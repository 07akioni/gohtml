package lexer

import "testing"

func TestFindTextEnd(t *testing.T) {
	got := findTextEnd("01234<6789", 0)
	if got != 5 {
		t.Errorf("number after <, %v", got)
	}
	got = findTextEnd("01234<!789", 0)
	if got != 5 {
		t.Errorf("! after <, %v", got)
	}
	got = findTextEnd("01234</789", 0)
	if got != 5 {
		t.Errorf("/ after <, %v", got)
	}
	got = findTextEnd("01234< 789", 0)
	if got != 10 {
		t.Errorf("space after <, %v", got)
	}
}
