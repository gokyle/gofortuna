/*
   package fortuna implements the Fortuna PRNG designed by Niels
   Ferguson, Bruce Schneier, and Yadayoshi Kohno. The PRNG is
   described in the book _Cryptography Engineering_, by the same
   authors (see pages 142-160). This implementation uses AES-256
   and as the underlying PRF and SHA-256 as the underlying PRG.

   The Fortuna type provided by this package contains the actual
   PRNG; clients should use one of the provided sources (or write
   their own) in order to add entropy to the PRNG.

   The book describes an alternative implementation in which a
   separate accumulator thread performs the hashing; this implementation
   takes the standard approach.

   The documentation for AddRandomEvent contains notes for writing
   new sources of random events to feed the PRNG.

   The book also recommends that the PRNG's seed file be updated
   regularly; at the very least, at shutdown with an update every
   ten minutes recommended.
*/
package fortuna
