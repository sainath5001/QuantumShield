package quantum

// AssessmentNotes are short, normative bullets attached to risk assessments.
func AssessmentNotes() []string {
	return []string{
		"Shor's algorithm (ECC discrete log) is the primary long-term concern for exposed secp256k1 ECDSA public keys under a hypothetical cryptographically relevant quantum computer (CRQC).",
		"Grover's algorithm offers a quadratic speedup for unstructured search and affects symmetric/hash security margins differently; it is not the same catastrophic break of ECDSA that Shor poses when the public key is known.",
		"Theoretical vulnerability under a future CRQC is not a practical attack with currently available quantum computers against Ethereum secp256k1 keys.",
		"Q-Guard does not claim that any quantum computer has broken an Ethereum key, and does not provide timelines or compromise probabilities.",
	}
}

// ReportExplanation returns structured educational copy for security reports.
func ReportExplanation(exposed bool) (shor, grover, practical string, bullets []string) {
	shor = "Shor's algorithm can solve elliptic-curve discrete logarithms in polynomial time on a sufficiently large fault-tolerant quantum computer. For Ethereum EOAs, that means a known secp256k1 public key could allow recovery of the corresponding private key under those assumptions."
	grover = "Grover's algorithm provides a quadratic speedup for unstructured search. It reduces idealized brute-force margins for symmetric cryptography and some hash preimage settings, but it does not replace Shor as the main narrative for ECDSA public-key exposure."
	practical = "No currently available quantum computer is known to break Ethereum secp256k1 keys. Q-Guard distinguishes theoretical long-term vulnerability of exposed keys from practical present-day attacks."
	bullets = AssessmentNotes()
	if exposed {
		bullets = append([]string{
			"This wallet's secp256k1 public key was observed in the examined scope; under a future CRQC, that key material is a higher-priority harvest-now/attack-later target than an address with no observed exposure.",
		}, bullets...)
	} else {
		bullets = append([]string{
			"No public key was confirmed exposed in the examined scope; the address-hash barrier still applies for that key material within this scope's limitations.",
		}, bullets...)
	}
	return shor, grover, practical, bullets
}
