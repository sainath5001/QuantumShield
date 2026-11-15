package cryptography

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

const (
	curveName   = "secp256k1"
	schemeName  = "ecdsa"
	evidenceRPC = "ethereum_rpc_transaction"
)

// RecoverOptions controls recovery verification.
type RecoverOptions struct {
	ChainID        *big.Int
	ExpectedSender common.Address // if zero, use types.Sender result as expected
	BlockNumber    *uint64
	Now            func() time.Time
}

// RecoverPublicKey recovers and verifies the secp256k1 public key for a signed transaction.
// It uses go-ethereum's canonical Signer.Hash as the signing digest.
func RecoverPublicKey(tx *types.Transaction, opts RecoverOptions) domain.ExposureEvidence {
	now := time.Now().UTC()
	if opts.Now != nil {
		now = opts.Now().UTC()
	}
	ev := domain.ExposureEvidence{
		Curve:              curveName,
		SignatureScheme:    schemeName,
		TransactionHash:    tx.Hash(),
		BlockNumber:        opts.BlockNumber,
		EvidenceSource:     evidenceRPC,
		AnalyzedAt:         now,
		SenderVerification: domain.SenderNotChecked,
	}
	txType := tx.Type()
	ev.TransactionType = &txType

	if opts.ChainID == nil || opts.ChainID.Sign() <= 0 {
		ev.RecoveryStatus = domain.RecoveryInconclusive
		ev.Detail = "chain ID is required for canonical signer selection"
		return ev
	}

	if !IsSupportedTxType(tx.Type()) {
		ev.RecoveryStatus = domain.RecoveryUnsupportedType
		ev.Detail = fmt.Sprintf("unsupported transaction type %d", tx.Type())
		return ev
	}

	signer := types.LatestSignerForChainID(opts.ChainID)

	// Canonical signing digest — must not use the signed transaction identifying hash.
	// Unprotected legacy txs are signed over the Homestead digest, not the EIP-155 digest.
	digest, err := canonicalSigningDigest(tx, opts.ChainID)
	if err != nil {
		ev.RecoveryStatus = mapSigError(err)
		ev.Detail = err.Error()
		return ev
	}
	ev.SigningDigestHex = digest.Hex()

	v, r, s := tx.RawSignatureValues()
	if r == nil || s == nil || v == nil {
		ev.RecoveryStatus = domain.RecoveryMissingSig
		ev.Detail = "transaction signature values are missing"
		return ev
	}
	if r.Sign() == 0 && s.Sign() == 0 {
		ev.RecoveryStatus = domain.RecoveryMissingSig
		ev.Detail = "transaction signature values are zero"
		return ev
	}

	plainV, err := plainRecoveryV(tx, v, opts.ChainID)
	if err != nil {
		ev.RecoveryStatus = mapSigError(err)
		ev.Detail = err.Error()
		return ev
	}

	sig, err := encodeSignature(r, s, plainV)
	if err != nil {
		ev.RecoveryStatus = domain.RecoveryInvalidSig
		ev.Detail = err.Error()
		return ev
	}

	pubBytes, err := crypto.Ecrecover(digest[:], sig)
	if err != nil {
		ev.RecoveryStatus = domain.RecoveryFailed
		ev.Detail = fmt.Sprintf("ecrecover failed: %v", err)
		return ev
	}
	if len(pubBytes) != 65 || pubBytes[0] != 0x04 {
		ev.RecoveryStatus = domain.RecoveryFailed
		ev.Detail = "ecrecover returned an unexpected public key encoding"
		return ev
	}

	pub, err := crypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		ev.RecoveryStatus = domain.RecoveryFailed
		ev.Detail = fmt.Sprintf("public key unmarshal failed: %v", err)
		return ev
	}

	derived := crypto.PubkeyToAddress(*pub)
	ev.DerivedSender = &derived
	ev.RecoveredPublicKeyHex = hexutil.Encode(pubBytes)
	ev.PublicKeyEncoding = domain.PubkeyUncompressedSEC1
	ev.PublicKeyFingerprint = PublicKeyFingerprint(pubBytes)

	// Cross-check against go-ethereum Sender (canonical address recovery).
	sender, err := types.Sender(signer, tx)
	if err != nil {
		ev.RecoveryStatus = mapSigError(err)
		ev.Detail = fmt.Sprintf("types.Sender failed after ecrecover: %v", err)
		ev.SenderVerification = domain.SenderNotChecked
		return ev
	}
	if sender != derived {
		ev.RecoveryStatus = domain.RecoveryInconclusive
		ev.Detail = "internal inconsistency: Ecrecover-derived address differs from types.Sender"
		ev.SenderVerification = domain.SenderMismatch
		return ev
	}

	expected := opts.ExpectedSender
	if expected == (common.Address{}) {
		expected = sender
	}
	ev.ExpectedSender = expected
	ev.WalletAddress = expected

	if derived != expected {
		ev.RecoveryStatus = domain.RecoverySenderMismatch
		ev.SenderVerification = domain.SenderMismatch
		ev.Detail = "recovered public key does not derive to the expected sender address"
		return ev
	}

	ev.RecoveryStatus = domain.RecoveryConfirmed
	ev.SenderVerification = domain.SenderVerified
	ev.Detail = "public key recovered and sender verified via go-ethereum canonical digest"
	return ev
}

// plainRecoveryV normalizes transaction V / yParity to Homestead-style 27/28 for recoverPlain.
func plainRecoveryV(tx *types.Transaction, v *big.Int, chainID *big.Int) (*big.Int, error) {
	switch tx.Type() {
	case types.LegacyTxType:
		if !tx.Protected() {
			// Homestead / Frontier: V is already 27 or 28.
			if v.Cmp(big.NewInt(27)) != 0 && v.Cmp(big.NewInt(28)) != 0 {
				return nil, fmt.Errorf("%w: unprotected legacy V=%s", errInvalidSig, v.String())
			}
			return new(big.Int).Set(v), nil
		}
		// EIP-155: V = chainID*2 + 35|36 → plain = V - chainID*2 - 8 → 27|28
		if tx.ChainId().Cmp(chainID) != 0 {
			return nil, fmt.Errorf("%w: tx chain id %s != configured %s", errInvalidChain, tx.ChainId(), chainID)
		}
		chainIDMul := new(big.Int).Mul(chainID, big.NewInt(2))
		plain := new(big.Int).Sub(v, chainIDMul)
		plain.Sub(plain, big.NewInt(8))
		if plain.Cmp(big.NewInt(27)) != 0 && plain.Cmp(big.NewInt(28)) != 0 {
			return nil, fmt.Errorf("%w: EIP-155 plain V=%s from raw V=%s", errInvalidSig, plain.String(), v.String())
		}
		return plain, nil
	case types.AccessListTxType, types.DynamicFeeTxType, types.BlobTxType:
		// Typed transactions store yParity in {0,1}.
		if v.Cmp(big.NewInt(0)) != 0 && v.Cmp(big.NewInt(1)) != 0 {
			return nil, fmt.Errorf("%w: typed tx yParity V=%s", errInvalidSig, v.String())
		}
		if tx.ChainId().Cmp(chainID) != 0 {
			return nil, fmt.Errorf("%w: tx chain id %s != configured %s", errInvalidChain, tx.ChainId(), chainID)
		}
		return new(big.Int).Add(v, big.NewInt(27)), nil
	default:
		return nil, fmt.Errorf("%w: type %d", errUnsupportedType, tx.Type())
	}
}

func encodeSignature(r, s, plainV *big.Int) ([]byte, error) {
	if plainV.BitLen() > 8 {
		return nil, fmt.Errorf("%w: recovery id too large", errInvalidSig)
	}
	recID := byte(plainV.Uint64() - 27)
	if !crypto.ValidateSignatureValues(recID, r, s, true) {
		return nil, fmt.Errorf("%w: ValidateSignatureValues rejected r/s/v", errInvalidSig)
	}
	rb, sb := r.Bytes(), s.Bytes()
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(rb):32], rb)
	copy(sig[64-len(sb):64], sb)
	sig[64] = recID
	return sig, nil
}

func mapSigError(err error) domain.RecoveryStatus {
	if err == nil {
		return domain.RecoveryInconclusive
	}
	if errors.Is(err, types.ErrTxTypeNotSupported) || errors.Is(err, errUnsupportedType) {
		return domain.RecoveryUnsupportedType
	}
	if errors.Is(err, types.ErrInvalidSig) || errors.Is(err, errInvalidSig) {
		return domain.RecoveryInvalidSig
	}
	if errors.Is(err, types.ErrInvalidChainId) || errors.Is(err, errInvalidChain) {
		return domain.RecoveryInconclusive
	}
	return domain.RecoveryFailed
}

var (
	errInvalidSig      = errors.New("invalid signature")
	errInvalidChain    = errors.New("invalid chain id")
	errUnsupportedType = errors.New("unsupported transaction type")
)

// PublicKeyFingerprint returns a short non-reversible fingerprint for logs/UI lists.
func PublicKeyFingerprint(uncompressedSEC1 []byte) string {
	sum := crypto.Keccak256(uncompressedSEC1)
	return hexutil.Encode(sum[:8])
}

// DeriveAddressFromUncompressed derives an Ethereum address from a SEC1 uncompressed pubkey.
func DeriveAddressFromUncompressed(pub *ecdsa.PublicKey) common.Address {
	return crypto.PubkeyToAddress(*pub)
}

// SigningDigest returns the canonical digest that must be ECDSA-signed for tx on chainID.
// This is not the signed transaction's identifying hash.
func SigningDigest(tx *types.Transaction, chainID *big.Int) (common.Hash, error) {
	if chainID == nil || chainID.Sign() <= 0 {
		return common.Hash{}, fmt.Errorf("chain ID required")
	}
	return canonicalSigningDigest(tx, chainID)
}

// canonicalSigningDigest mirrors go-ethereum Sender digest selection per tx type.
func canonicalSigningDigest(tx *types.Transaction, chainID *big.Int) (common.Hash, error) {
	if !IsSupportedTxType(tx.Type()) {
		return common.Hash{}, fmt.Errorf("%w: type %d", errUnsupportedType, tx.Type())
	}
	switch tx.Type() {
	case types.LegacyTxType:
		if !tx.Protected() {
			return types.HomesteadSigner{}.Hash(tx), nil
		}
		return types.NewEIP155Signer(chainID).Hash(tx), nil
	case types.AccessListTxType:
		return types.NewEIP2930Signer(chainID).Hash(tx), nil
	case types.DynamicFeeTxType:
		return types.NewLondonSigner(chainID).Hash(tx), nil
	case types.BlobTxType:
		return types.NewCancunSigner(chainID).Hash(tx), nil
	default:
		return common.Hash{}, fmt.Errorf("%w: type %d", errUnsupportedType, tx.Type())
	}
}

// IsSupportedTxType reports whether recovery is implemented for the envelope type.
func IsSupportedTxType(t byte) bool {
	switch t {
	case types.LegacyTxType, types.AccessListTxType, types.DynamicFeeTxType, types.BlobTxType:
		return true
	default:
		return false
	}
}
