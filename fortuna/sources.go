package fortuna

import "fmt"

// SourceChannel provides an interface to a PRNG that reads random
// events from a channel and adds them to the PRNG for entropy. The
// source number s should be used by the application to identify
// this particular source.
type SourceChannel struct {
	rng *Fortuna
	s   byte
	i   int
	In  chan []byte // In receives incoming random events.
	Out chan error  // Out sends outgoing errors.
}

// NewSourceChannel initialises a new channel source. This is
// required to properly initialise one. The source parameter should
// contain the source number. The rng must already be initialised,
// and the channel source must be started before it can be used.
func NewSourceChannel(rng *Fortuna, source byte) *SourceChannel {
	if rng == nil {
		return nil
	} else if !rng.Initialised() {
		return nil
	}

	return &SourceChannel{
		rng: rng,
		s:   source,
		i:   0,
	}
}

// Start the channel source, setting up the channel sender and
// receiver.
func (cs *SourceChannel) Start(buf int) {
	cs.In = make(chan []byte, buf)
	cs.Out = make(chan error, buf)
	go func() {
		for {
			e, ok := <-cs.In
			if !ok {
				return
			}
			err := cs.rng.AddRandomEvent(cs.s, cs.i, e)
			if err != nil {
				cs.Out <- err
			}
			cs.i = (cs.i + 1) % len(cs.rng.pools)
		}
	}()
}

// Stop halts the channel source, closing the channels.
func (cs *SourceChannel) Stop() {
	if cs.In != nil {
		close(cs.In)
		cs.In = nil
	}

	if cs.Out != nil {
		close(cs.Out)
		cs.Out = nil
	}
}

// SourceWriter provides an io.Writer source for adding events to
// the PRNG.
type SourceWriter struct {
	rng *Fortuna
	s   byte
	i   int
}

// NewSourceWriter intialises a new io.Writer source. This is
// required to properly intialise the source. The PRNG provided
// must already be initialised; the source parameter is used to
// identify the source to the host system.
func NewSourceWriter(rng *Fortuna, source byte) *SourceWriter {
	if rng == nil || !rng.Initialised() {
		return nil
	}

	return &SourceWriter{
		rng: rng,
		s:   source,
	}
}

// Write adds the byte slice as entropy to the pools in the PRNG.
func (sw *SourceWriter) Write(p []byte) (int, error) {
	if p == nil {
		return 0, nil
	}
	pp := p
	k := len(p) / MaxEventSize
	if len(p)%MaxEventSize != 0 {
		k++
	}

	for i := 0; i < k; i++ {
		wrsz := MaxEventSize
		if i == (k-1) && ((len(pp) % MaxEventSize) != 0) {
			wrsz = len(pp) % MaxEventSize
		}
		err := sw.rng.AddRandomEvent(sw.s, sw.i, pp[:wrsz])
		sw.i = (sw.i + 1) % len(sw.rng.pools)
		if err != nil {
			fmt.Printf("%d, wrsz: %d, len(p): %d\n", i, wrsz, len(p))
			return len(p) - len(pp), err
		}
		pp = pp[wrsz:]
	}
	return len(p), nil
}
