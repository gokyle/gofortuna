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
const (
	maxBlocks    int = 65536
	newKeyBlocks int = 2 // len(rngKey) / aes.BlocksSize
)

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

// New initialises a new Fortuna generator context. This is required
// to properly initialise a new generator instance.
func NewGenerator() *Generator {
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
func (g *Generator) Reseed(s string) {
	h := sha256.New()
	h.Write(g.key[:])
	h.Write([]byte(s))
	key := h.Sum(nil)
	copy(g.key[:], key)
	zero(key)
	h.Reset()
	incCounter(g.ctr)
}

func (g *Generator) generateBlocks(k int) (r []byte, err error) {
	r = make([]byte, 0, k*aes.BlockSize)

	c, err := aes.NewCipher(g.key[:])
	if err != nil {
		return
	}

	for i := 0; i < k; i++ {
		block := make([]byte, aes.BlockSize)
		c.Encrypt(block, g.ctr[:])
		r = append(r, block...)
		incCounter(g.ctr)
	}
	return
}

// Write performs the same operation as Reseed, but allows the
// generator to be used as an io.Writer.
func (g *Generator) Write(bs []byte) (int, error) {
	h := sha256.New()
	h.Write(g.key[:])
	h.Write(bs)
	key := h.Sum(nil)
	copy(g.key[:], key)
	zero(key)
	h.Reset()
	incCounter(g.ctr)
	return len(bs), nil
}

func (g *Generator) blockGenerate(p []byte, k int) error {
	r, err := g.generateBlocks(k)
	if err != nil {
		return err
	}
	nk, err := g.generateBlocks(newKeyBlocks)
	if err != nil {
		return err
	}
	copy(p, r)
	copy(g.key[:], nk)
	return nil
}

// Read presents the generator as an io.Reader, and is used to read
// random data from the generator.
func (g *Generator) Read(p []byte) (int, error) {
	if p == nil {
		return 0, nil
	}

	pl := len(p)
	k := pl / aes.BlockSize
	if (pl % aes.BlockSize) != 0 {
		k++
	}
	gr := k / maxBlocks
	if (k % maxBlocks) != 0 {
		gr++
	}

	pp := p
	for i := 0; i < gr; i++ {
		blks := maxBlocks
		if i == (gr-1) && ((k % maxBlocks) != 0) {
			blks = k % maxBlocks
		}
		err := g.blockGenerate(pp, blks)
		if err != nil {
			return i * aes.BlockSize, err
		}
		if i < (gr - 1) {
			pp = pp[blks*aes.BlockSize:]
		}
	}
	return pl, nil
}
