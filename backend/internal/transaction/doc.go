// Package transaction will decode Ethereum transaction envelopes and rebuild
// signing digests for supported types (legacy, EIP-2930, EIP-1559, EIP-4844).
//
// Normalized history models currently live in package domain. Signing-digest
// recovery is not implemented yet.
package transaction
