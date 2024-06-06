package lcg

import (
	"encoding/base64"
	"encoding/binary"
	"strconv"
	"testing"
)

// Logging is the slowest part about this
func TestLCG(t *testing.T) {
	seed := GenRandomUIntNotPrime()

	t.Log("Seed: " + strconv.FormatInt(int64(seed), 10))

	lcg := New(seed)
	if lcg.seed != seed {
		t.Errorf("Expected seed to be 0, got %d", lcg.seed)
	}
	if lcg.a != 6364136223846793005 {
		t.Errorf("Expected a to be 6364136223846793005, got %d", lcg.a)
	}
	if lcg.c != 1 {
		t.Errorf("Expected c to be 1, got %d", lcg.c)
	}

	map1 := make(map[int64]bool)

	for i := 0; i < 1000000; i++ {
		n := lcg.Next()
		in := int64(n)
		if map1[in] {
			t.Log(strconv.Itoa(i) + " Number " + strconv.FormatInt(in, 10) + "  already generated")
			t.Fail()
		}
		map1[in] = true
		b := make([]byte, binary.MaxVarintLen64)
		_ = binary.PutVarint(b, in)
		t.Log(strconv.Itoa(i) + " Generated unique number: " + strconv.FormatInt(in, 10) + " " + base64.URLEncoding.EncodeToString(b))
	}

	for i := 0; i < 1000000; i++ {
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
		t.Log(strconv.Itoa(i) + " Generated unique number: " + strconv.FormatInt(in, 10))
	}
}
