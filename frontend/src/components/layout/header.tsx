"use client";

import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { getApiBaseUrl } from "@/lib/api/client";
import { getHealth } from "@/lib/api/qguard";

export function Header({ title, subtitle }: { title: string; subtitle?: string }) {
  const [apiOk, setApiOk] = useState<"checking" | "ok" | "down">("checking");

  useEffect(() => {
    let cancelled = false;
    getHealth()
      .then(() => {
        if (!cancelled) setApiOk("ok");
      })
      .catch(() => {
        if (!cancelled) setApiOk("down");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <header className="flex flex-col gap-3 border-b border-border px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 className="text-xl font-semibold tracking-tight text-foreground">{title}</h1>
        {subtitle ? <p className="mt-1 text-sm text-muted-foreground">{subtitle}</p> : null}
      </div>
      <div className="flex flex-wrap items-center gap-2 text-xs">
        <Badge variant="outline">Ethereum Mainnet</Badge>
        <Badge variant={apiOk === "ok" ? "cyan" : apiOk === "down" ? "elevated" : "outline"}>
          API {apiOk === "checking" ? "checking…" : apiOk === "ok" ? "reachable" : "unreachable"}
        </Badge>
        <span className="hidden font-mono text-muted-foreground md:inline">{getApiBaseUrl()}</span>
      </div>
    </header>
  );
}
