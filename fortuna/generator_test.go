package fortuna

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestReseed(t *testing.T) {
	expected := "8df823ade13d19bb8d73973193c50cf02559afcaf460397d1a459e1d3466941c"
	expectedCtr := &rngCounter{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	seed := "initial state"
	g := NewGenerator()
	g.Reseed(seed)
	if fmt.Sprintf("%x", g.key[:]) != expected {
		fmt.Fprintf(os.Stderr, "fortuna: key failure on reseed\n")
		t.FailNow()
	} else if !bytes.Equal(expectedCtr[:], g.ctr[:]) {
		fmt.Fprintf(os.Stderr, "fortuna: counter failure on reseed\n")
		fmt.Fprintf(os.Stderr, "\t counter: %x\n", g.ctr[:])
		fmt.Fprintf(os.Stderr, "\texpected: %x\n", expectedCtr[:])
		t.FailNow()
	}
}

func TestGenerateBlocks(t *testing.T) {
	expected := "fcdfb28a3fb0a1527dca5c083fac33fd6c591974bdfaa1a757bd7a85bc6db717"
	g := NewGenerator()
	g.Reseed("initial state")
	r, err := g.generateBlocks(2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}
	if fmt.Sprintf("%x", r) != expected {
		fmt.Fprintf(os.Stderr, "fortuna: bad blocks in generateBlocks\n")
		t.FailNow()
	}
}

func TestBadGenerateBlocks(t *testing.T) {
	expected := "fcdfb28a3fb0a1527dca5c083fac33fd6c591974bdfaa1a757bd7a85bc6db717"
	g := NewGenerator()
	g.Reseed("initial state 2")
	r, err := g.generateBlocks(2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	}
	if fmt.Sprintf("%x", r) == expected {
		fmt.Fprintf(os.Stderr, "fortuna: invalid blocks in generateBlocks\n")
		t.FailNow()
	}
}

func TestWrite(t *testing.T) {
	expected := "8df823ade13d19bb8d73973193c50cf02559afcaf460397d1a459e1d3466941c"
	expectedCtr := &rngCounter{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	seed := []byte("initial state")
	g := NewGenerator()
	n, err := g.Write(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != len(seed) {
		fmt.Fprintf(os.Stderr, "fortuna: bad length on write\n")
		t.FailNow()
	} else if fmt.Sprintf("%x", g.key[:]) != expected {
		fmt.Fprintf(os.Stderr, "fortuna: key failure on write\n")
		t.FailNow()
	} else if !bytes.Equal(expectedCtr[:], g.ctr[:]) {
		fmt.Fprintf(os.Stderr, "fortuna: counter failure on reseed\n")
		fmt.Fprintf(os.Stderr, "\t counter: %x\n", g.ctr[:])
		fmt.Fprintf(os.Stderr, "\texpected: %x\n", expectedCtr[:])
		t.FailNow()
	}
}

func TestRead(t *testing.T) {
	seed := []byte("initial state")
	expected := "fcdfb28a3fb0a1527dca5c083fac33fd6c591974bdfaa1a7"
	expectedKey := "23fddd8d1c7d9a2615b60ccfc40441165b443f37cea7452fe8d9544d1b1d2fca"
	g := NewGenerator()
	n, err := g.Write(seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != len(seed) {
		fmt.Fprintf(os.Stderr, "fortuna: bad length on write\n")
		t.FailNow()
	}

	r := make([]byte, 24)
	n, err = g.Read(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		t.FailNow()
	} else if n != len(r) {
		fmt.Fprintf(os.Stderr, "fortuna: short read\n")
		t.FailNow()
	}

	if fmt.Sprintf("%x", r) != expected {
		fmt.Fprintf(os.Stderr, "fortuna: invalid output\n")
		fmt.Fprintf(os.Stderr, "\t  actual: %x\n", r)
		fmt.Fprintf(os.Stderr, "\texpected: %s\n", expected)
		t.FailNow()
	} else if fmt.Sprintf("%x", g.key[:]) != expectedKey {
		fmt.Fprintf(os.Stderr, "fortuna: invalid key after read\n")
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
		fmt.Fprintf(os.Stderr, "fortuna: bad length on write\n")
		t.FailNow()
	}

	var p []byte
	n, err = g.Read(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		t.FailNow()
	} else if n != 0 {
		fmt.Fprintf(os.Stderr, "fortuna: read should have returned no data\n")
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
