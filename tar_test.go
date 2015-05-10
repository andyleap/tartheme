// test.go
package tartheme

import "testing"

func TestTarLoad(t *testing.T) {

	tt, _ := Load("theme.tar")

	if tt.Assets["static/test.css"].Name != "static/test.css" {
		t.Logf("%q != %q\n", tt.Assets["static/test.css"].Name, "static/test.css")
	}

}
