package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "correct-horse-battery-staple" {
		t.Fatal("password was not hashed")
	}
	if !CheckPassword(hash, "correct-horse-battery-staple") {
		t.Fatal("expected password to verify")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected wrong password to fail")
	}
}
