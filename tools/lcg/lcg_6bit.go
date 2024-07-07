package lcg

const MASK uint64 = (1 << 48) - 1 // Mask for 48 bits

// LCG48 is a linear congruential generator to generate pseudo-random numbers
// and fill out a 2^48 byte space without hitting the same number twice.
type LCG48 struct {
	seed uint64
	a    uint64
	c    uint64
}

func New48(seed uint64) *LCG48 {
	return &LCG48{
		seed: seed & MASK,
		c:    ONE,
		a:    25214903917,
	}
}

func (l *LCG48) Next() uint64 {
	l.seed = (l.a*l.seed + l.c) & MASK
	return l.seed
}

func (l *LCG48) Skip(skip int64) {
	delta := uint64(skip) & MASK
	a := l.a
	c := l.c
	a_next := ONE
	c_next := ZERO
	for delta > 0 {
		if (delta & ONE) != ZERO {
			a_next = (a_next * a) & MASK
			c_next = (c_next*a + c) & MASK
		}
		c = ((a + ONE) * c) & MASK
		a = (a * a) & MASK
		delta >>= ONE

		if delta > MASK {
			// If remaining delta is larger than our 48-bit space,
			// we can skip the rest as it will wrap around
			break
		}
	}
	l.seed = (a_next*l.seed + c_next) & MASK
}
