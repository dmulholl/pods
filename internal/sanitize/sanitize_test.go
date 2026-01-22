package sanitize

import "testing"

func TestFilename(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "test 1",
			input: `abc`,
			want:  `abc`,
		},
		{
			name:  "test 2",
			input: `a\\\\*b|/?/??//<<>>:"c`,
			want:  `a-b-c`,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			got := Filename(testcase.input)
			if got != testcase.want {
				t.Errorf("wanted %q, got %q", testcase.want, got)
			}
		})
	}
}
