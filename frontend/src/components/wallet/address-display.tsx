"use client";

import { Check, Copy } from "lucide-react";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn, shortenAddress } from "@/lib/utils";

export function AddressDisplay({
  address,
  className,
  compact = false,
}: {
  address: string;
  className?: string;
  compact?: boolean;
}) {
  const [copied, setCopied] = useState(false);
  const shown = compact ? shortenAddress(address, 6) : address;

  async function copy() {
    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* ignore */
    }
  }

  return (
    <div className={cn("inline-flex items-center gap-2", className)}>
      <code className="rounded bg-muted px-2 py-1 font-mono text-xs text-cyan sm:text-sm">{shown}</code>
      <Button type="button" variant="ghost" size="icon" className="h-8 w-8" onClick={copy} aria-label="Copy address">
        {copied ? <Check className="h-3.5 w-3.5 text-lower" /> : <Copy className="h-3.5 w-3.5" />}
      </Button>
      <Badge variant="outline">Mainnet</Badge>
    </div>
  );
}
