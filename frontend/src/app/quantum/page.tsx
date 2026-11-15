"use client";

import { Header } from "@/components/layout/header";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const sections = [
  {
    title: "Quantum fundamentals",
    accent: "quantum" as const,
    body: [
      "Qubits hold complex amplitudes; measurement yields classical bits.",
      "Superposition and entanglement enable algorithmic speedups for specific problems — not magic decryption of arbitrary data.",
      "Circuits compose unitary gates followed by measurement and classical post-processing.",
    ],
  },
  {
    title: "Shor's algorithm",
    accent: "cyan" as const,
    body: [
      "Period finding + classical post-processing can factor integers / break discrete-log schemes at cryptographically large scales — in theory, on a fault-tolerant quantum computer.",
      "Q-Guard's Python module demonstrates a toy N=15 example only. That is not an Ethereum private-key derivation.",
      "secp256k1 discrete-log instances used by EOAs are far beyond educational simulators.",
    ],
  },
  {
    title: "Grover's algorithm",
    accent: "cyan" as const,
    body: [
      "Quadratic speedup for unstructured search. It does not 'directly break Ethereum wallets' the way Shor threatens ECC with known public keys.",
      "Educational demos use tiny search spaces (≤4 qubits).",
    ],
  },
  {
    title: "Post-quantum cryptography",
    accent: "quantum" as const,
    body: [
      "ML-DSA (signatures), ML-KEM (key encapsulation — not a signature), SLH-DSA (hash-based signatures).",
      "Replacing Ethereum EOA signatures requires protocol and wallet ecosystem support — not a notebook experiment.",
      "Using PQC in an unrelated demo does not make a scanned wallet quantum-resistant.",
    ],
  },
];

export default function QuantumLabPage() {
  return (
    <div>
      <Header
        title="Quantum Research Lab"
        subtitle="Educational material. Simulation results are never invented in this UI."
      />
      <div className="space-y-6 p-6">
        <Card className="border-quantum/30">
          <CardHeader>
            <div className="flex flex-wrap items-center gap-2">
              <CardTitle>Experiment runner status</CardTitle>
              <Badge variant="quantum">Integration pending</Badge>
            </div>
            <CardDescription>
              The Python package <code className="text-quantum">qguard_quantum</code> exposes CLI and
              callable interfaces (<code className="text-quantum">qguard-quantum shor|grover|pqc</code>).
              A FastAPI sidecar is optional and not required for MVP. This page does not display fake
              measurement histograms.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>When integrated, this lab will call documented experiment entrypoints and render returned JSON only.</p>
            <p>
              Locally today: <code className="text-foreground">cd quantum-engine && source .venv/bin/activate && qguard-quantum info</code>
            </p>
            <p className="rounded-md border border-border bg-muted/30 p-3 text-xs">
              Hard boundary: quantum demo outcomes must never feed wallet scan evidence or risk classification.
            </p>
          </CardContent>
        </Card>

        <div className="grid gap-4 lg:grid-cols-2">
          {sections.map((s) => (
            <Card key={s.title}>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  {s.title}
                  <Badge variant={s.accent === "quantum" ? "quantum" : "cyan"}>Education</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="list-disc space-y-2 pl-5 text-sm text-muted-foreground">
                  {s.body.map((line) => (
                    <li key={line}>{line}</li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          ))}
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Circuits & simulation results</CardTitle>
            <CardDescription>No placeholder charts. Run the Python module for real AerSimulator output.</CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Circuit metadata (qubit count, depth, gate tallies, measurement counts) is produced by the
            educational engine with fixed seeds for reproducibility. Wire-up path for a future UI action:
            invoke <code className="text-cyan">qguard_quantum.interface.run_experiment</code> via a
            documented backend bridge — not by fabricating counts in React.
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
