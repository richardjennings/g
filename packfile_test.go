package g

import "testing"

func TestLookup(t *testing.T) {
	Configure()
	sha, err := ShaFromHexString("d208f34d505adb8914e5a5c577c6db8359173b4e")
	//sha, err := ShaFromHexString("ebf679b4719cb6df2e6d9f8a471353abae08cf98")
	if err != nil {
		t.Fatal(err)
	}
	lookupInPackfiles(sha)
}
