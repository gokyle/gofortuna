package tunafish

import (
	"code.google.com/p/go.crypto/sha3"
	"errors"
	"io"
	"io/ioutil"
	"sync"
	"time"
)

// MinPoolSize stores the number of bytes that will trigger a reseed.
// The ReseedDelay prevents reseed events from occuring too quickly.
var (
	MinPoolSize int64 = 32
	ReseedDelay       = 100 * time.Millisecond
)

// MaxEventSize is the limit to the amount of data that can be sent
// in an event.
const MaxEventSize = 32

// PoolSize contains the number of pools used by the PRNG.
const PoolSize = 32

// SeedFileLength is the number of bytes that should be present in
// the seed file.
const SeedFileLength = 64

var (
	ErrNotSeeded    = errors.New("tunafish: PRNG not seeded yet")
	ErrInvalidEvent = errors.New("tunafish: invalid random event")
	ErrInvalidSeed  = errors.New("tunafish: invalid seed")
)

type pool struct {
	hash    []byte
	written int64
	sync.Mutex
}

type reseedTime struct {
	time.Time
	sync.Mutex
}

type Tunafish struct {
	initialised bool
	pools       *[32]*pool
	counter     uint32
	g           *Generator
	lastReseed  *reseedTime
}

// Initialised returns true if the rng is initialised.
func (rng *Tunafish) Initialised() bool {
	if rng == nil {
		return false
	}
	return rng.initialised
}

// New sets up a new Fortuna PRNG; it is required for ensuring that
// the PRNG is properly initialised.
func New() *Tunafish {
	rng := &Tunafish{
		pools:      new([32]*pool),
		g:          NewGenerator(),
		lastReseed: &reseedTime{},
	}

	for i := range rng.pools {
		rng.pools[i] = &pool{
			hash: []byte{},
		}
	}

	rng.initialised = true
	return rng
}

func (rng *Tunafish) mustReseed() bool {
	rng.pools[0].Lock()
	poolReseed := rng.pools[0].written >= MinPoolSize
	rng.pools[0].Unlock()

	rng.lastReseed.Lock()
	reseed := rng.lastReseed.Time.Add(ReseedDelay)
	rng.lastReseed.Unlock()
	return poolReseed && time.Now().After(reseed)
}

func (rng *Tunafish) reseed() {
	rng.counter++
	s := []byte{}

	for i := 0; i < len(rng.pools); i++ {
		if ((1 << uint32(i)) | rng.counter) != 0 {
			rng.pools[i].Lock()
			h := sha3.NewKeccak256()
			h.Write(rng.pools[i].hash)
			s = append(s, h.Sum(nil)...)
			rng.pools[i].hash = []byte{}
			rng.pools[i].Unlock()
		}
	}
	rng.g.Write(s)
	rng.lastReseed.Lock()
	rng.lastReseed.Time = time.Now()
	rng.lastReseed.Unlock()
}

func (rng *Tunafish) Read(p []byte) (int, error) {
	if rng.mustReseed() {
		rng.reseed()
	}

	if rng.counter == 0 {
		return 0, ErrNotSeeded
	}

	if p == nil {
		return 0, nil
	}

	return rng.g.Read(p)
}

// AddRandomEvent should be called by sources to add random events
// to the PRNG; it takes a source identifier, a pool number, and a
// random event. Sources should cycle through pools, evenly
// distributing events over the entire set of pools; the Fortuna
// designers specify that this should be done "in a round-robin
// fashion." The choice of a source identifier is up to the host
// application.
func (rng *Tunafish) AddRandomEvent(s byte, i int, e []byte) error {
	if e == nil || len(e) == 0 || len(e) > MaxEventSize {
		return ErrInvalidEvent
	}

	if i < 0 || i > len(rng.pools) {
		return ErrInvalidEvent
	}

	rng.pools[i].Lock()
	rng.pools[i].hash = append(rng.pools[i].hash, s)
	rng.pools[i].hash = append(rng.pools[i].hash, byte(len(e)))
	rng.pools[i].hash = append(rng.pools[i].hash, e...)
	rng.pools[i].written += int64(len(e) + 2)
	rng.pools[i].Unlock()
	return nil
}

// Seed dumps a byte slice containing a seed that may be used to
// restore the PRNG's state.
func (rng *Tunafish) Seed() ([]byte, error) {
	var p = make([]byte, SeedFileLength)
	_, err := io.ReadFull(rng, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// WriteSeed writes a seed to a file; this should be used for
// restoring the PRNG state later.
func (rng *Tunafish) WriteSeed(filename string) error {
	seed, err := rng.Seed()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, seed, 0600)
}

// UpdateSeed reads a seed from a file and updates the seed file
// with new random data.
func (rng *Tunafish) UpdateSeed(filename string) error {
	seed, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	} else if len(seed) != SeedFileLength {
		return ErrInvalidSeed
	}

	rng.g.Write(seed)
	return rng.WriteSeed(filename)
}

// ReadSeed reseeds the PRNG with a seed that is expected to have
// been read from a seed file.
func (rng *Tunafish) ReadSeed(p []byte) error {
	if len(p) != SeedFileLength {
		return ErrInvalidSeed
	}
	rng.g.Write(p)
	return nil
}

// FromSeed creates a new PRNG instance from the seed file. This
// can be used to start an RNG on start up.
func FromSeed(filename string) (*Tunafish, error) {
	seed, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	} else if len(seed) != SeedFileLength {
		return nil, ErrInvalidSeed
	}

	rng := New()
	rng.g.Write(seed)
	return rng, nil
}
