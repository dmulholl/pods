package sanitize

import "testing"

func TestFilename(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no change",
			input: `abc`,
			want:  `abc`,
		},
		{
			name:  "trim spaces",
			input: ` abc  `,
			want:  `abc`,
		},
		{
			name:  "individual illegal characters",
			input: `?a/b<c:`,
			want:  `a-b-c`,
		},
		{
			name:  "multiple illegal characters",
			input: `a\\\\*b|/?/??//<<>>:"c`,
			want:  `a-b-c`,
		},
		{
			name:  "illegal characters and spaces",
			input: ` a\\\\*b|/?/??//<<>>:"c  `,
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
