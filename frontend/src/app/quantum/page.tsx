"use client";

import { useCallback, useEffect, useState } from "react";
import { Header } from "@/components/layout/header";
import { ErrorState } from "@/components/states/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { humanizeQuantumError } from "@/lib/api/errors";
import {
  getPqcExplainer,
  getQuantumApiBaseUrl,
  getQuantumHealth,
  listQuantumDemos,
  runGroverDemo,
  runShorDemo,
  type QuantumDemoMeta,
} from "@/lib/api/quantum";

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
  const [serviceOk, setServiceOk] = useState<"checking" | "ok" | "down">("checking");
  const [demos, setDemos] = useState<QuantumDemoMeta[]>([]);
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<Record<string, unknown> | null>(null);

  const refresh = useCallback(async () => {
    setServiceOk("checking");
    setError(null);
    try {
      await getQuantumHealth();
      const listed = await listQuantumDemos();
      setDemos(listed.demos);
      setServiceOk("ok");
    } catch (err) {
      setServiceOk("down");
      setDemos([]);
      setError(humanizeQuantumError(err));
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function run(kind: "shor" | "grover" | "pqc") {
    setBusy(kind);
    setError(null);
    setResult(null);
    try {
      if (kind === "shor") setResult(await runShorDemo());
      else if (kind === "grover") setResult(await runGroverDemo());
      else setResult(await getPqcExplainer());
    } catch (err) {
      setError(humanizeQuantumError(err));
    } finally {
      setBusy(null);
    }
  }

  return (
    <div>
      <Header
        title="Quantum Research Lab"
        subtitle="Educational experiments via the Python sidecar. Results are never invented in this UI."
      />
      <div className="space-y-6 p-6">
        <Card className="border-quantum/30">
          <CardHeader>
            <div className="flex flex-wrap items-center gap-2">
              <CardTitle>Experiment runner</CardTitle>
              <Badge variant={serviceOk === "ok" ? "quantum" : serviceOk === "down" ? "elevated" : "outline"}>
                {serviceOk === "checking" ? "Checking…" : serviceOk === "ok" ? "Service reachable" : "Service down"}
              </Badge>
            </div>
            <CardDescription>
              Thin FastAPI wrapper around <code className="text-quantum">qguard_quantum.interface</code> at{" "}
              <code className="text-quantum">{getQuantumApiBaseUrl()}</code>. Hard boundary: outcomes never feed
              wallet scan evidence.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => void refresh()} disabled={busy !== null}>
                Refresh status
              </Button>
              <Button
                type="button"
                variant="cyan"
                size="sm"
                disabled={serviceOk !== "ok" || busy !== null}
                onClick={() => void run("shor")}
              >
                {busy === "shor" ? "Running Shor…" : "Run Shor (N=15)"}
              </Button>
              <Button
                type="button"
                variant="cyan"
                size="sm"
                disabled={serviceOk !== "ok" || busy !== null}
                onClick={() => void run("grover")}
              >
                {busy === "grover" ? "Running Grover…" : "Run Grover (2 qubits)"}
              </Button>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                disabled={serviceOk !== "ok" || busy !== null}
                onClick={() => void run("pqc")}
              >
                {busy === "pqc" ? "Loading…" : "Load PQC overview"}
              </Button>
            </div>
            {demos.length ? (
              <ul className="list-disc pl-5 text-xs text-muted-foreground">
                {demos.map((d) => (
                  <li key={d.id}>
                    <span className="text-foreground">{d.name}</span> — {d.path}
                    {d.notes ? ` (${d.notes})` : ""}
                  </li>
                ))}
              </ul>
            ) : null}
            {error ? <ErrorState title="Quantum service error" message={error} onRetry={() => void refresh()} /> : null}
            {result ? (
              <pre className="max-h-96 overflow-auto rounded-md border border-border bg-muted/30 p-3 text-xs leading-relaxed text-muted-foreground">
                {JSON.stringify(result, null, 2)}
              </pre>
            ) : null}
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
      </div>
    </div>
  );
}
