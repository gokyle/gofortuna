package fortuna

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestNilRNG(t *testing.T) {
	var rng *Fortuna
	if rng.Initialised() {
		fmt.Fprintf(os.Stderr, "fortuna: PRNG should not have reported it was initialised\n")
		t.FailNow()
	}
}

func TestNotSeeded(t *testing.T) {
	var p []byte
	rng := New(nil)
	if _, err := rng.Read(p); err == nil {
		fmt.Fprintf(os.Stderr, "fortuna: PRNG should report it is not seeded")
		t.FailNow()
	}
}

func TestPRNGEmptyRead(t *testing.T) {
	var p []byte
	rng := New(nil)
	rng.reseed()

	n, err := rng.Read(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != 0 {
		fmt.Fprintf(os.Stderr, "fortuna: no data should have been read\n")
		t.FailNow()
	}
}

func TestInvalidEvents(t *testing.T) {
	var p = make([]byte, 1)
	rng := New(nil)
	err := rng.AddRandomEvent(0, 33, p)
	if err != ErrInvalidEvent {
		fmt.Fprintf(os.Stderr, "fortuna: random event should be invalid\n")
		t.FailNow()
	}

	err = rng.AddRandomEvent(0, -1, p)
	if err != ErrInvalidEvent {
		fmt.Fprintf(os.Stderr, "fortuna: random event should be invalid\n")
		t.FailNow()
	}

	p = nil
	err = rng.AddRandomEvent(0, 0, p)
	if err != ErrInvalidEvent {
		fmt.Fprintf(os.Stderr, "fortuna: random event should be invalid\n")
		t.FailNow()
	}

	p = make([]byte, MaxEventSize+1)
	err = rng.AddRandomEvent(0, 0, p)
	if err != ErrInvalidEvent {
		fmt.Fprintf(os.Stderr, "fortuna: random event should be invalid\n")
		t.FailNow()
	}
}

var seed []byte

func TestSeed(t *testing.T) {
	rng := New(nil)
	sw := NewSourceWriter(rng, 0)

	_, err := rng.Seed()
	if err != ErrNotSeeded {
		fmt.Fprintf(os.Stderr, "fortuna: PRNG seed() should fail for unseeded PRNG")
		t.FailNow()
	}

	f, err := os.Open("/dev/zero")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}

	io.CopyN(sw, f, 4096)
	seed, err = rng.Seed()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if len(seed) != SeedFileLength {
		fmt.Fprintf(os.Stderr, "fortuna: bad seed file length\n")
		t.FailNow()
	}
}

func TestReadSeed(t *testing.T) {
	rng := New(nil)
	err := rng.ReadSeed(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}

	seed = nil
	err = rng.ReadSeed(seed)
	if err == nil {
		fmt.Fprintf(os.Stderr, "fortuna: ReadSeed should fail\n")
		t.FailNow()
	}

	seed = make([]byte, SeedFileLength-1)
	if err = rng.ReadSeed(seed); err == nil {
		fmt.Fprintf(os.Stderr, "fortuna: ReadSeed should fail\n")
		t.FailNow()
	}
}
