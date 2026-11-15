// Package cryptography performs ECDSA public-key recovery and address verification
// for Ethereum transactions using go-ethereum primitives exclusively.
//
// Algorithm (EOA transactions):
//  1. Select types.LatestSignerForChainID(chainID) (Cancun signer: legacy through EIP-4844).
//  2. Obtain the canonical signing digest via signer.Hash(tx) — never tx.Hash() alone.
//  3. Read RawSignatureValues (v, r, s) and normalize the recovery id per transaction type.
//  4. Recover the secp256k1 public key with crypto.Ecrecover / crypto.SigToPub.
//  5. Derive the address with crypto.PubkeyToAddress and verify against types.Sender and tx.from.
//
// Confirmed exposure requires successful recovery AND sender verification.
// This package never accepts private keys, seeds, or passwords, and never signs transactions.
package cryptography
