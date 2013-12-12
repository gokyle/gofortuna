package tunafish

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestReseed(t *testing.T) {
	expected := "a3987997c3fae1735e98b76098392a893c938111a442baa28606af33d6b13ca8"
	expectedCtr := &rngCounter{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	seed := "initial state"
	g := NewGenerator()
	g.Reseed(seed)
	if fmt.Sprintf("%x", g.key[:]) != expected {
		fmt.Fprintf(os.Stderr, "tunafish: key failure on reseed\n")
		t.FailNow()
	} else if !bytes.Equal(expectedCtr[:], g.ctr[:]) {
		fmt.Fprintf(os.Stderr, "tunafish: counter failure on reseed\n")
		fmt.Fprintf(os.Stderr, "\t counter: %x\n", g.ctr[:])
		fmt.Fprintf(os.Stderr, "\texpected: %x\n", expectedCtr[:])
		t.FailNow()
	}
}

func TestGenerateBlocks(t *testing.T) {
	expected := "8600a1b8594e89a03423691c1b446e7779b6df1c39019c30375af3f07a145d8c"
	g := NewGenerator()
	g.Reseed("initial state")
	r, err := g.generateBlocks(2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}
	if fmt.Sprintf("%x", r) != expected {
		fmt.Fprintf(os.Stderr, "tunafish: bad blocks in generateBlocks\n")
		t.FailNow()
	}
}

func TestBadGenerateBlocks(t *testing.T) {
	expected := "8600a1b8594e89a03423691c1b446e7779b6df1c39019c30375af3f07a145d8c"
	g := NewGenerator()
	g.Reseed("initial state 2")
	r, err := g.generateBlocks(2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}
	if fmt.Sprintf("%x", r) == expected {
		fmt.Fprintf(os.Stderr, "tunafish: invalid blocks in generateBlocks\n")
		t.FailNow()
	}
}

func TestWrite(t *testing.T) {
	expected := "a3987997c3fae1735e98b76098392a893c938111a442baa28606af33d6b13ca8"
	expectedCtr := &rngCounter{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	seed := []byte("initial state")
	g := NewGenerator()
	n, err := g.Write(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != len(seed) {
		fmt.Fprintf(os.Stderr, "tunafish: bad length on write\n")
		t.FailNow()
	} else if fmt.Sprintf("%x", g.key[:]) != expected {
		fmt.Fprintf(os.Stderr, "tunafish: key failure on write\n")
		t.FailNow()
	} else if !bytes.Equal(expectedCtr[:], g.ctr[:]) {
		fmt.Fprintf(os.Stderr, "tunafish: counter failure on reseed\n")
		fmt.Fprintf(os.Stderr, "\t counter: %x\n", g.ctr[:])
		fmt.Fprintf(os.Stderr, "\texpected: %x\n", expectedCtr[:])
		t.FailNow()
	}
}

func TestRead(t *testing.T) {
	seed := []byte("initial state")
	expected := "8600a1b8594e89a03423691c1b446e7779b6df1c39019c30"
	expectedKey := "edc54d00473f22ccd7dc77ddde2dfa74fb421afac09cf261eeb86880606ed284"
	g := NewGenerator()
	n, err := g.Write(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != len(seed) {
		fmt.Fprintf(os.Stderr, "tunafish: bad length on write\n")
		t.FailNow()
	}

	r := make([]byte, 24)
	n, err = g.Read(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		t.FailNow()
	} else if n != len(r) {
		fmt.Fprintf(os.Stderr, "tunafish: short read\n")
		t.FailNow()
	}

	if fmt.Sprintf("%x", r) != expected {
		fmt.Fprintf(os.Stderr, "tunafish: invalid output\n")
		fmt.Fprintf(os.Stderr, "\t  actual: %x\n", r)
		fmt.Fprintf(os.Stderr, "\texpected: %s\n", expected)
		t.FailNow()
	} else if fmt.Sprintf("%x", g.key[:]) != expectedKey {
		fmt.Fprintf(os.Stderr, "tunafish: invalid key after read\n")
		t.FailNow()
	}
}

func TestEmptyRead(t *testing.T) {
	seed := []byte("initial state")
	g := NewGenerator()
	n, err := g.Write(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != len(seed) {
		fmt.Fprintf(os.Stderr, "tunafish: bad length on write\n")
		t.FailNow()
	}

	var p []byte
	n, err = g.Read(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != 0 {
		fmt.Fprintf(os.Stderr, "tunafish: read should have returned no data\n")
		t.FailNow()
	}
}

func BenchmarkGeneratorRead4k(b *testing.B) {
	g := NewGenerator()
	g.Reseed("initial state")

	r := make([]byte, 4096)
	for i := 0; i < b.N; i++ {
		_, err := g.Read(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			b.FailNow()
		}
	}
}

func BenchmarkGeneratorRead4M(b *testing.B) {
	g := NewGenerator()
	g.Reseed("initial state")
	r := make([]byte, 4*1024*1024)
	for i := 0; i < b.N; i++ {
		_, err := g.Read(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			b.FailNow()
		}
	}
}
