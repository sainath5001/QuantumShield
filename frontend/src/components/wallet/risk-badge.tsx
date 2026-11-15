import { Badge } from "@/components/ui/badge";
import type { RiskClassification, WalletExposureFinding } from "@/lib/models/types";

const riskMap: Record<
  RiskClassification,
  { label: string; variant: "elevated" | "lower" | "unknown" | "review" }
> = {
  ELEVATED_EXPOSURE: { label: "Elevated exposure", variant: "elevated" },
  LOWER_EXPOSURE: { label: "Lower exposure", variant: "lower" },
  UNKNOWN: { label: "Unknown", variant: "unknown" },
  REQUIRES_REVIEW: { label: "Requires review", variant: "review" },
};

export function RiskBadge({ classification }: { classification?: RiskClassification | null }) {
  if (!classification) return <Badge variant="outline">No assessment</Badge>;
  const meta = riskMap[classification];
  return <Badge variant={meta.variant}>{meta.label}</Badge>;
}

export function FindingBadge({ finding }: { finding?: WalletExposureFinding | null }) {
  if (!finding) return <Badge variant="outline">No finding</Badge>;
  const labels: Record<WalletExposureFinding, string> = {
    public_key_exposed: "Public key exposed (confirmed)",
    no_exposure_observed_in_scope: "No exposure observed in scope",
    insufficient_evidence: "Insufficient evidence",
    requires_manual_review: "Requires manual review",
  };
  const variant =
    finding === "public_key_exposed"
      ? "elevated"
      : finding === "no_exposure_observed_in_scope"
        ? "lower"
        : finding === "requires_manual_review"
          ? "review"
          : "unknown";
  return <Badge variant={variant}>{labels[finding]}</Badge>;
}
