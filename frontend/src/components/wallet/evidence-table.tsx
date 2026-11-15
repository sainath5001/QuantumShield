import { EmptyState } from "@/components/states/empty-state";
import { Badge } from "@/components/ui/badge";
import type { ExposureEvidence } from "@/lib/models/types";
import { shortenAddress } from "@/lib/utils";

export function EvidenceTable({ evidence }: { evidence: ExposureEvidence[] }) {
  if (!evidence.length) {
    return (
      <EmptyState
        title="No evidence rows"
        description="Confirmed public-key recovery evidence will appear here when available."
      />
    );
  }

  return (
    <div className="overflow-x-auto rounded-lg border border-border">
      <table className="w-full min-w-[720px] text-left text-sm">
        <thead className="bg-muted/40 text-xs uppercase tracking-wide text-muted-foreground">
          <tr>
            <th className="px-3 py-2 font-medium">Tx hash</th>
            <th className="px-3 py-2 font-medium">Block</th>
            <th className="px-3 py-2 font-medium">Recovery</th>
            <th className="px-3 py-2 font-medium">Sender check</th>
            <th className="px-3 py-2 font-medium">Fingerprint</th>
            <th className="px-3 py-2 font-medium">Scheme</th>
          </tr>
        </thead>
        <tbody>
          {evidence.map((row) => (
            <tr key={`${row.transaction_hash}-${row.analyzed_at}`} className="border-t border-border">
              <td className="px-3 py-2 font-mono text-xs text-cyan">
                {shortenAddress(row.transaction_hash, 8)}
              </td>
              <td className="px-3 py-2 tabular-nums text-muted-foreground">
                {row.block_number ?? "—"}
              </td>
              <td className="px-3 py-2">
                <Badge variant={row.recovery_status === "confirmed" ? "elevated" : "outline"}>
                  {row.recovery_status}
                </Badge>
              </td>
              <td className="px-3 py-2">
                <Badge variant={row.sender_verification_status === "verified" ? "lower" : "outline"}>
                  {row.sender_verification_status}
                </Badge>
              </td>
              <td className="px-3 py-2 font-mono text-xs">
                {row.public_key_fingerprint ?? "—"}
              </td>
              <td className="px-3 py-2 text-xs text-muted-foreground">
                {row.curve} / {row.signature_scheme}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
