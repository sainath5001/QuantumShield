import { Header } from "@/components/layout/header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const topics = [
  {
    title: "Ethereum ECDSA",
    body: `Ethereum externally owned accounts (EOAs) authorize transactions with ECDSA signatures over the secp256k1 curve. A signature is produced with a private key; anyone can verify it with the corresponding public key. Address derivation hashes the public key — the address alone does not reveal the full public key until a signature is published on-chain (or elsewhere).`,
  },
  {
    title: "Public-key exposure",
    body: `When a wallet sends a transaction, the signature can allow recovery of the secp256k1 public key. Q-Guard confirms exposure only when recovery succeeds and the derived address matches the analyzed wallet. Exposure is not the same as knowing the private key today, and it is not evidence of a quantum attack.`,
  },
  {
    title: "Quantum computing threats",
    body: `Shor's algorithm threatens schemes whose security relies on factoring or discrete logarithms (including elliptic-curve schemes) if a large fault-tolerant quantum computer exists. Grover's algorithm offers a quadratic speedup for unstructured search and is a different, usually weaker, concern for symmetric cryptography key sizes. Timelines are uncertain; Q-Guard does not estimate them.`,
  },
  {
    title: "Post-quantum cryptography",
    body: `NIST standardized ML-DSA (signatures), ML-KEM (key encapsulation), and SLH-DSA (hash-based signatures). ML-KEM is not a digital signature scheme. These algorithms aim to resist known classical and quantum attacks under different hardness assumptions (lattices, hashes).`,
  },
  {
    title: "Ethereum migration challenges",
    body: `Replacing EOA secp256k1 signatures is not a library swap. Transaction formats, address schemes, wallets, hardware devices, RPC tooling, and smart-contract recovery flows all assume today's cryptography. Account abstraction and new transaction types are part of broader protocol/ecosystem migration — far beyond a research demo.`,
  },
];

export default function LearnPage() {
  return (
    <div>
      <Header
        title="Research & learn"
        subtitle="Beginner-friendly explanations without sacrificing technical accuracy."
      />
      <div className="mx-auto max-w-3xl space-y-4 p-6">
        {topics.map((t) => (
          <Card key={t.title}>
            <CardHeader>
              <CardTitle>{t.title}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm leading-relaxed text-muted-foreground">{t.body}</p>
            </CardContent>
          </Card>
        ))}
        <Card>
          <CardHeader>
            <CardTitle>How this relates to Q-Guard</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>
              The scanner answers a classical question: has this Mainnet address shown verifiable public-key
              exposure within scoped history?
            </p>
            <p>
              The Quantum Lab and this Learn section explain why that question matters for long-term
              cryptography planning — without claiming scanned wallets are quantum-safe or already broken.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
