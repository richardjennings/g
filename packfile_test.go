package g

import "testing"

func TestLookup(t *testing.T) {
	Configure()
	sha, err := ShaFromHexString("d208f34d505adb8914e5a5c577c6db8359173b4e")
	//sha, err := ShaFromHexString("f209394e4150235bf75037bba9bf49c671edf091")
	if err != nil {
		t.Fatal(err)
	}
	Lookup(sha)
}
