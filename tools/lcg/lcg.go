package lcg

const ZERO uint64 = 0
const ONE uint64 = 1

// LCG is a linear congruential generator to generate pseudo-random numbers
// and fill out a 2^64 byte space without hitting the same number twice.
type LCG struct {
	seed uint64
	a    uint64
	c    uint64
}

func New(seed uint64) *LCG {
	return &LCG{
		seed: seed,
		c:    1,
		a:    6364136223846793005,
	}
}

func (l *LCG) Next() uint64 {
	l.seed = l.a*l.seed + l.c
	return l.seed
}

func (l *LCG) Skip(skip int64) {
	/*
		Signed argument - skip forward as well as backward

		The algorithm here to determine the parameters used to skip ahead is
		described in the paper F. Brown, "Random Number Generation with Arbitrary Stride,"
		Trans. Am. Nucl. Soc. (Nov. 1994). This algorithm is able to skip ahead in
		O(log2(N)) operations instead of O(N). It computes parameters
		A and C which can then be used to find x_N = A*x_0 + C mod 2^M.
	*/

	nskip := uint64(skip)

	a := l.a
	c := l.c

	a_next := ONE
	c_next := ZERO

	for nskip > 0 {
		if (nskip & ONE) != ZERO {
			a_next = a_next * a
			c_next = c_next*a + c
		}
		c = (a + ONE) * c
		a = a * a

		nskip = nskip >> ONE
	}

	l.seed = a_next*l.seed + c_next
}
