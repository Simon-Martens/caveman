package lcg

const ZERO uint64 = 0
const ONE uint64 = 1
const MAX uint64 = 0x0000FFFFFFFFFFFF

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
		  -> F. Brown, "Random Number Generation with Arbitrary Stride," 1994

			Complexity: O(log2(N)), not O(N).

			It computes parameters A and C which can then be used to find
			x_N = A*x_0 + C mod 2^M.
	*/

	delta := uint64(skip)

	a := l.a
	c := l.c

	a_next := ONE
	c_next := ZERO

	for delta > 0 {
		if (delta & ONE) != ZERO {
			a_next = a_next * a
			c_next = c_next*a + c
		}
		c = (a + ONE) * c
		a = a * a

		delta = delta >> ONE
	}

	l.seed = a_next*l.seed + c_next
}
