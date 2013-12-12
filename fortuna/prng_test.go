package fortuna

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
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
	rng := New()
	if _, err := rng.Read(p); err == nil {
		fmt.Fprintf(os.Stderr, "fortuna: PRNG should report it is not seeded")
		t.FailNow()
	}
}

func TestPRNGEmptyRead(t *testing.T) {
	var p []byte
	rng := New()
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
	rng := New()
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
	rng := New()
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
	rng := New()
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

func TestSeedFiles(t *testing.T) {
	rng := New()
	sw := NewSourceWriter(rng, 0)

	f, err := os.Open("/dev/zero")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}

	io.CopyN(sw, f, 4096)
	outFile := "test.seed"
	defer os.Remove(outFile)
	if err = rng.WriteSeed(outFile); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if err = rng.UpdateSeed(outFile); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}

	partialSeed := seed[2:]
	if err = ioutil.WriteFile(outFile, partialSeed, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if err = rng.UpdateSeed(outFile); err == nil {
		fmt.Fprintf(os.Stderr, "fortuna: PRNG should not accept an invalid seed\n")
		t.FailNow()
	} else if err = rng.UpdateSeed("invalid.seed"); err == nil {
		fmt.Fprintf(os.Stderr, "fortuna: PRNG should not accept a non-existent seed\n")
		t.FailNow()
	}

	if _, err = FromSeed(outFile); err == nil {
		fmt.Fprintln(os.Stderr, "fortuna: restoring from seed shuold fail with short seed", err)
		t.FailNow()
	} else if _, err = FromSeed("invalid.seed"); err == nil {
		fmt.Fprintf(os.Stderr, "fortuna: restoring from seed should fail with non-existent seed\n")
		t.FailNow()
	} else if err = rng.WriteSeed(outFile); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if _, err = FromSeed(outFile); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}
}

func BenchmarkFortunaRead4096(b *testing.B) {
	rng := New()
	sw := NewSourceWriter(rng, 1)
	n, err := io.CopyN(sw, rand.Reader, 4096)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		b.FailNow()
	} else if n != 4096 {
		fmt.Fprintln(os.Stderr, "fortuna: failed to seed PRNG")
		b.FailNow()
	}

	var p = make([]byte, 4096)
	for i := 0; i < b.N; i++ {
		_, err = rng.Read(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			b.FailNow()
		}

	}
}
