package bkrs

import "testing"

func TestCleanBkrsLine(t *testing.T) {
	tests := []struct {
		term     string
		expected string
	}{
		{" \n[m1]сесть на мель[/m]", "сесть на мель"},
		{"располагать\\[ся\\]", "располагать[ся]"},
		{
			" [m1]1) 《诗‧小雅》篇名。[/m][m1]2) 喻丈夫之恩惠。[/m][m1]3) 浓重的露水。[/m]",
			"1) 《诗‧小雅》篇名。\n2) 喻丈夫之恩惠。\n3) 浓重的露水。",
		},
	}
	cleaner := makeCleaner()

	for _, testCase := range tests {
		result := CleanBkrsLine(testCase.term, cleaner)

		if result != testCase.expected {
			t.Errorf("Incorrect result. Expect %s, got %s", testCase.expected, result)
		}
	}
}
