// Written in 2012 by Dmitry Chestnykh.
//
// To the extent possible under law, the author have dedicated all copyright
// and related and neighboring rights to this software to the public domain
// worldwide. This software is distributed without any warranty.
// http://creativecommons.org/publicdomain/zero/1.0/

// Package siphash implements SipHash-2-4, a fast short-input PRF
// created by Jean-Philippe Aumasson and Daniel J. Bernstein.
package sip

const (
	// The block size of hash algorithm in bytes.
	BlockSize = 8
	// The size of hash output in bytes.
	Size = 8
)

type digest struct {
	k0, k1 uint64 // two parts of key
}

// New returns a new hash.Hash64 computing SipHash-2-4 with 16-byte key.
func New(key []byte) *digest {
	d := new(digest)

	d.k0 = uint64(key[0]) | uint64(key[1])<<8 | uint64(key[2])<<16 | uint64(key[3])<<24 |
		uint64(key[4])<<32 | uint64(key[5])<<40 | uint64(key[6])<<48 | uint64(key[7])<<56

	d.k1 = uint64(key[8]) | uint64(key[9])<<8 | uint64(key[10])<<16 | uint64(key[11])<<24 |
		uint64(key[12])<<32 | uint64(key[13])<<40 | uint64(key[14])<<48 | uint64(key[15])<<56

	return d
}

func (d *digest) Hash(p []byte) uint64 {
	return Hash(d.k0, d.k1, p)
}
