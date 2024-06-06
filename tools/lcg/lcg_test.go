package lcg

import (
	"strconv"
	"testing"
)

func TestLCG(t *testing.T) {
	seed := GenRandomUIntNotPrime()
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

	for i := 0; i < 10000000; i++ {
		n := lcg.Next()
		in := int64(n)
		if map1[in] {
			t.Log(strconv.Itoa(i) + " Number " + strconv.FormatInt(in, 10) + "  already generated")
			t.Fail()
		}
		map1[in] = true
		t.Log(strconv.Itoa(i) + " Generated unique number: " + strconv.FormatInt(in, 10))
	}
}
