package fortuna

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
)

type (
	rngKey     [32]byte
	rngCounter [16]byte
)

const MaxRead int = 1048576

var ErrReadTooLarge = errors.New("fortuna: can't provide requested number of bytes")

// Generator represents the underlying PRG used by the Fortuna PRNG.
type Generator struct {
	key *rngKey
	ctr *rngCounter
}

func incCounter(ctr *rngCounter) {
	l := len(ctr)
	for i := 0; i < l; i++ {
		if ctr[i]++; ctr[i] != 0 {
			return
		}
	}
}

// New initialises a new Fortuna PRNG context. This is required to
// properly initialise a new Fortunate instance.
func New() *Generator {
	return &Generator{
		key: new(rngKey),
		ctr: new(rngCounter),
	}
}

func zero(bs []byte) {
	if bs == nil {
		return
	}
	bsl := len(bs)
	for i := 0; i < bsl; i++ {
		bs[i] ^= bs[i]
	}
}

// Reseed reseeds the generator with the given arbitrary input.
func (rng *Generator) Reseed(s string) {
	h := sha256.New()
	h.Write(rng.key[:])
	h.Write([]byte(s))
	key := h.Sum(nil)
	copy(rng.key[:], key)
	zero(key)
	h.Reset()
	incCounter(rng.ctr)
}

func (rng *Generator) generateBlocks(k int) (r []byte, err error) {
	r = make([]byte, 0, k*aes.BlockSize)

	c, err := aes.NewCipher(rng.key[:])
	if err != nil {
		return
	}

	for i := 0; i < k; i++ {
		block := make([]byte, aes.BlockSize)
		c.Encrypt(block, rng.ctr[:])
		r = append(r, block...)
		incCounter(rng.ctr)
	}
	return
}

// Write performs the same operation as Reseed, but allows the
// generator to be used as an io.Writer.
func (rng *Generator) Write(bs []byte) (int, error) {
	h := sha256.New()
	h.Write(rng.key[:])
	h.Write(bs)
	key := h.Sum(nil)
	copy(rng.key[:], key)
	zero(key)
	h.Reset()
	incCounter(rng.ctr)
	return len(bs), nil
}

// TODO(kyle): remove read length limitation by wrapping Read around
// as many invocations of generateBlocks as required.

// Read presents the generator as an io.Reader, and is used to read
// random data from the generator.
func (rng *Generator) Read(bs []byte) (int, error) {
	if bs == nil {
		return 0, nil
	} else if len(bs) > MaxRead {
		return 0, ErrReadTooLarge
	}

	k := len(bs) / 16
	if (len(bs) % 16) != 0 {
		k++
	}

	r, err := rng.generateBlocks(k)
	if err != nil {
		return 0, err
	}

	nk, err := rng.generateBlocks(2)
	if err != nil {
		return 0, err
	}
	copy(bs, r)
	copy(rng.key[:], nk)
	return len(bs), nil
}
