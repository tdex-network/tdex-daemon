package permissions

import "testing"

func TestValidatePermissions(t *testing.T) {
	if err := validatePermissions(); err != nil {
		t.Fatal(err)
	}
}
