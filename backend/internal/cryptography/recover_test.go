package cryptography

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

// Well-known Hardhat/Anvil account #0 — used only as a reproducible test vector, not as scan evidence.
const testPrivHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func testKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := crypto.HexToECDSA(testPrivHex)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func testAddr(t *testing.T) common.Address {
	t.Helper()
	return crypto.PubkeyToAddress(testKey(t).PublicKey)
}

func chainID() *big.Int { return big.NewInt(1) }

func signLegacyEIP155(t *testing.T, key *ecdsa.PrivateKey) *types.Transaction {
	t.Helper()
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	tx := types.NewTransaction(1, to, big.NewInt(1), 21000, big.NewInt(1), nil)
	signer := types.LatestSignerForChainID(chainID())
	signed, err := types.SignTx(tx, signer, key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func signLegacyHomestead(t *testing.T, key *ecdsa.PrivateKey) *types.Transaction {
	t.Helper()
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	tx := types.NewTransaction(2, to, big.NewInt(1), 21000, big.NewInt(1), nil)
	signer := types.HomesteadSigner{}
	signed, err := types.SignTx(tx, signer, key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func signAccessList(t *testing.T, key *ecdsa.PrivateKey) *types.Transaction {
	t.Helper()
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	inner := &types.AccessListTx{
		ChainID:  chainID(),
		Nonce:    3,
		To:       &to,
		Value:    big.NewInt(1),
		Gas:      21000,
		GasPrice: big.NewInt(1),
		Data:     nil,
	}
	tx := types.NewTx(inner)
	signer := types.LatestSignerForChainID(chainID())
	signed, err := types.SignTx(tx, signer, key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func signDynamicFee(t *testing.T, key *ecdsa.PrivateKey) *types.Transaction {
	t.Helper()
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	inner := &types.DynamicFeeTx{
		ChainID:   chainID(),
		Nonce:     4,
		To:        &to,
		Value:     big.NewInt(1),
		Gas:       21000,
		GasFeeCap: big.NewInt(2),
		GasTipCap: big.NewInt(1),
		Data:      nil,
	}
	tx := types.NewTx(inner)
	signer := types.LatestSignerForChainID(chainID())
	signed, err := types.SignTx(tx, signer, key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func signBlob(t *testing.T, key *ecdsa.PrivateKey) *types.Transaction {
	t.Helper()
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	inner := &types.BlobTx{
		ChainID:    uint256.MustFromBig(chainID()),
		Nonce:      5,
		To:         to,
		Value:      uint256.NewInt(0),
		Gas:        21000,
		GasFeeCap:  uint256.NewInt(2),
		GasTipCap:  uint256.NewInt(1),
		BlobFeeCap: uint256.NewInt(1),
		BlobHashes: []common.Hash{common.HexToHash("0x0100000000000000000000000000000000000000000000000000000000000000")},
	}
	tx := types.NewTx(inner)
	signer := types.LatestSignerForChainID(chainID())
	signed, err := types.SignTx(tx, signer, key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestRecoverAllSupportedTypes(t *testing.T) {
	key := testKey(t)
	addr := testAddr(t)
	cases := []struct {
		name string
		tx   *types.Transaction
		typ  byte
	}{
		{"legacy_eip155", signLegacyEIP155(t, key), types.LegacyTxType},
		{"legacy_homestead", signLegacyHomestead(t, key), types.LegacyTxType},
		{"eip2930", signAccessList(t, key), types.AccessListTxType},
		{"eip1559", signDynamicFee(t, key), types.DynamicFeeTxType},
		{"eip4844", signBlob(t, key), types.BlobTxType},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.tx.Type() != tc.typ {
				t.Fatalf("type = %d want %d", tc.tx.Type(), tc.typ)
			}
			ev := RecoverPublicKey(tc.tx, RecoverOptions{ChainID: chainID(), ExpectedSender: addr})
			if ev.RecoveryStatus != domain.RecoveryConfirmed {
				t.Fatalf("status = %s detail=%s", ev.RecoveryStatus, ev.Detail)
			}
			if ev.SenderVerification != domain.SenderVerified {
				t.Fatalf("sender verification = %s", ev.SenderVerification)
			}
			if ev.DerivedSender == nil || *ev.DerivedSender != addr {
				t.Fatalf("derived = %v", ev.DerivedSender)
			}
			if ev.RecoveredPublicKeyHex == "" {
				t.Fatal("missing pubkey")
			}
			// Compare against canonical Sender.
			sender, err := types.Sender(types.LatestSignerForChainID(chainID()), tc.tx)
			if err != nil {
				// Homestead unprotected uses HomesteadSigner
				sender, err = types.Sender(types.HomesteadSigner{}, tc.tx)
			}
			if err != nil {
				t.Fatal(err)
			}
			if sender != addr {
				t.Fatalf("canonical sender = %s", sender.Hex())
			}
			digest, err := SigningDigest(tc.tx, chainID())
			if err != nil {
				t.Fatal(err)
			}
			if digest.Hex() != ev.SigningDigestHex {
				t.Fatalf("digest mismatch")
			}
			// For typed txs, identifying hash must not be used as signing digest.
			if tc.typ != types.LegacyTxType && digest == tc.tx.Hash() {
				t.Fatal("typed tx signing digest must not equal identifying hash")
			}
		})
	}
}

func TestAddressDerivationMatchesPubkeyToAddress(t *testing.T) {
	key := testKey(t)
	want := crypto.PubkeyToAddress(key.PublicKey)
	got := DeriveAddressFromUncompressed(&key.PublicKey)
	if got != want {
		t.Fatalf("got %s want %s", got.Hex(), want.Hex())
	}
}

func TestSenderMismatch(t *testing.T) {
	key := testKey(t)
	tx := signDynamicFee(t, key)
	wrong := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	ev := RecoverPublicKey(tx, RecoverOptions{ChainID: chainID(), ExpectedSender: wrong})
	if ev.RecoveryStatus != domain.RecoverySenderMismatch {
		t.Fatalf("status = %s", ev.RecoveryStatus)
	}
	if ev.SenderVerification != domain.SenderMismatch {
		t.Fatalf("verification = %s", ev.SenderVerification)
	}
}

func TestInvalidSignatureValues(t *testing.T) {
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	inner := &types.DynamicFeeTx{
		ChainID:   chainID(),
		Nonce:     1,
		To:        &to,
		Value:     big.NewInt(0),
		Gas:       21000,
		GasFeeCap: big.NewInt(1),
		GasTipCap: big.NewInt(1),
		V:         big.NewInt(0),
		R:         big.NewInt(0),
		S:         big.NewInt(0),
	}
	tx := types.NewTx(inner)
	ev := RecoverPublicKey(tx, RecoverOptions{ChainID: chainID(), ExpectedSender: testAddr(t)})
	if ev.RecoveryStatus != domain.RecoveryMissingSig {
		t.Fatalf("status = %s detail=%s", ev.RecoveryStatus, ev.Detail)
	}
}

func TestMalformedRecoveryID(t *testing.T) {
	key := testKey(t)
	signed := signDynamicFee(t, key)
	v, r, s := signed.RawSignatureValues()
	// Rebuild with invalid yParity 2.
	to := *signed.To()
	bad := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID(),
		Nonce:     signed.Nonce(),
		To:        &to,
		Value:     signed.Value(),
		Gas:       signed.Gas(),
		GasFeeCap: signed.GasFeeCap(),
		GasTipCap: signed.GasTipCap(),
		Data:      signed.Data(),
		V:         big.NewInt(2),
		R:         new(big.Int).Set(r),
		S:         new(big.Int).Set(s),
	})
	_ = v
	ev := RecoverPublicKey(bad, RecoverOptions{ChainID: chainID(), ExpectedSender: testAddr(t)})
	if ev.RecoveryStatus != domain.RecoveryInvalidSig {
		t.Fatalf("status = %s detail=%s", ev.RecoveryStatus, ev.Detail)
	}
}

func TestIncorrectSignatureDoesNotConfirm(t *testing.T) {
	key := testKey(t)
	signed := signDynamicFee(t, key)
	v, r, s := signed.RawSignatureValues()
	to := *signed.To()
	// Flip s to break the signature while keeping structure.
	badS := new(big.Int).Xor(s, big.NewInt(1))
	bad := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID(),
		Nonce:     signed.Nonce(),
		To:        &to,
		Value:     signed.Value(),
		Gas:       signed.Gas(),
		GasFeeCap: signed.GasFeeCap(),
		GasTipCap: signed.GasTipCap(),
		Data:      signed.Data(),
		V:         new(big.Int).Set(v),
		R:         new(big.Int).Set(r),
		S:         badS,
	})
	ev := RecoverPublicKey(bad, RecoverOptions{ChainID: chainID(), ExpectedSender: testAddr(t)})
	if ev.RecoveryStatus == domain.RecoveryConfirmed {
		t.Fatal("broken signature must not confirm exposure")
	}
}

func TestMissingChainID(t *testing.T) {
	tx := signDynamicFee(t, testKey(t))
	ev := RecoverPublicKey(tx, RecoverOptions{ExpectedSender: testAddr(t)})
	if ev.RecoveryStatus != domain.RecoveryInconclusive {
		t.Fatalf("status = %s", ev.RecoveryStatus)
	}
}

func TestRecoveredPubkeyMatchesKnownKey(t *testing.T) {
	key := testKey(t)
	tx := signLegacyEIP155(t, key)
	ev := RecoverPublicKey(tx, RecoverOptions{ChainID: chainID(), ExpectedSender: testAddr(t)})
	if ev.RecoveryStatus != domain.RecoveryConfirmed {
		t.Fatal(ev.Detail)
	}
	want := crypto.FromECDSAPub(&key.PublicKey)
	got := common.FromHex(ev.RecoveredPublicKeyHex)
	if common.Bytes2Hex(want) != common.Bytes2Hex(got) {
		t.Fatalf("pubkey mismatch\nwant %x\ngot  %x", want, got)
	}
}
