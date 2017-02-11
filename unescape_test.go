package iislogs

import "testing"

func TestUnescape(t *testing.T) {
	test := []struct{ s, r string }{
		{"test", "test"},
		{"<%41>", "<A>"},
		{"<%3b>", "<;>"},
		{"Abc%3E%3A(%3b%2c)%3Def", "Abc>:(;,)=ef"},
		{"", ""},
		{"%41%2YTest", "AYTest"},
	}

	for _, c := range test {
		r := unescape(c.s)
		if r != c.r {
			t.Errorf("Expecting '%s', got '%s' (%v,%v)", c.r, r, []byte(c.r), []byte(r))
		}
	}
}
