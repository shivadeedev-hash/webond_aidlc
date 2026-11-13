package testdb

import "testing"

func TestResolve(t *testing.T) {
	cases := []struct {
		name      string
		dsn       string
		allow     bool
		wantDSN   bool
		wantError bool
	}{
		{"not configured", "", false, false, false},
		{"separate test db", "u:p@tcp(127.0.0.1:3306)/webond_test?parseTime=true", false, true, false},
		{"app db refused", "u:p@tcp(127.0.0.1:3306)/webond?parseTime=true", false, false, true},
		{"app db explicitly allowed", "u:p@tcp(127.0.0.1:3306)/webond", true, true, false},
		{"garbage", "not a dsn", false, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dsn, reason := Resolve(c.dsn, c.allow)
			if (dsn != "") != c.wantDSN {
				t.Errorf("dsn = %q, want non-empty=%v", dsn, c.wantDSN)
			}
			if (reason != "") != c.wantError {
				t.Errorf("reason = %q, want error=%v", reason, c.wantError)
			}
		})
	}
}
