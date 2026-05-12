package sqlsafe

import "testing"

func TestValidateAddsDefaultLimit(t *testing.T) {
	got, err := ValidateAndRewrite("SELECT * FROM customers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "SELECT * FROM customers LIMIT 100"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestValidateCapsLimit(t *testing.T) {
	got, err := ValidateAndRewrite("SELECT * FROM customers LIMIT 900")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "SELECT * FROM customers LIMIT 500"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestValidateRejectsDangerousSQL(t *testing.T) {
	cases := []string{
		"DELETE FROM customers",
		"DROP TABLE customers",
		"SELECT * FROM customers; SELECT * FROM orders",
		"PRAGMA table_info(customers)",
		"SELECT * FROM customers -- comment",
		"INSERT INTO customers(name) VALUES('x')",
	}
	for _, tc := range cases {
		if _, err := ValidateAndRewrite(tc); err == nil {
			t.Fatalf("expected %q to be rejected", tc)
		}
	}
}

func TestValidateAllowsWithSelect(t *testing.T) {
	got, err := ValidateAndRewrite("WITH totals AS (SELECT 1 AS n) SELECT n FROM totals")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected rewritten sql")
	}
}
