package lcg

import (
	"encoding/binary"
	"strconv"
	"testing"

	"github.com/Simon-Martens/caveman/tools/security"
)

// Logging is the slowest part about this
func TestLCG(t *testing.T) {
	seed := security.GenRandomUIntNotPrime()

	t.Log("Seed: " + strconv.FormatInt(int64(seed), 10))

	lcg := New48(seed)
	if lcg.seed != seed&MASK {
		t.Errorf("Expected seed to be 0, got %d", lcg.seed)
	}

	map1 := make(map[int64]bool)

	for i := 0; i < 10000000; i++ {
		n := lcg.Next()
		in := int64(n)
		if map1[in] {
			t.Log(strconv.Itoa(i) + " Number " + strconv.FormatInt(in, 10) + "  already generated")
			t.Fail()
		}
		map1[in] = true
		b := make([]byte, binary.MaxVarintLen64)
		_ = binary.PutVarint(b, in)
		//t.Log(strconv.Itoa(i) + " Generated unique number: " + strconv.FormatInt(in, 10) + " " + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b))
	}

	lcg.Skip(10000)

	for i := 0; i < 100; i++ {
		l := lcg.Next()
		in := int64(l)
		if l == 0 {
			t.Errorf("Expected l to be not 0, got %d", l)
		}

		lcg.Skip(-1)
		m := lcg.Next()
		if in != int64(m) {
			t.Errorf("Expected l to be equal to m, got %d and %d", l, m)
		}

		b := make([]byte, binary.MaxVarintLen64)
		_ = binary.PutVarint(b, in)
		// t.Log(strconv.Itoa(i) + " Generated unique number: " + strconv.FormatInt(in, 10) + " " + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b))
	}
}
