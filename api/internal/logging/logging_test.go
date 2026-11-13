package logging

import "testing"

func TestMaskEmail(t *testing.T) {
	// For local parts of length ≤ 3 only the first character is kept, so the
	// full local part is never revealed.
	cases := map[string]string{
		"investor@webond.co.th": "inv***@webond.co.th",
		"ab@x.com":              "a***@x.com",
		"a@x.com":               "a***@x.com",
		"not-an-email":          "***",
		"":                      "***",
	}
	for in, want := range cases {
		if got := MaskEmail(in); got != want {
			t.Errorf("MaskEmail(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMaskEmail_NeverLeaksFullLocalPart(t *testing.T) {
	// A long local part must be truncated to 3 chars + mask.
	got := MaskEmail("verylongname@example.com")
	if got != "ver***@example.com" {
		t.Errorf("MaskEmail leaked local part: %q", got)
	}
}
