package utils

import "testing"

func TestGenerateRandomString(t *testing.T) {
	b, err := GenerateRandomBytes(16)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)

	// Example: this will give us a 44 byte, base64 encoded output
	token, err := GenerateRandomString(32)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
}
