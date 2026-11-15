"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  BookOpen,
  FileText,
  FlaskConical,
  LayoutDashboard,
  ScanSearch,
  Shield,
} from "lucide-react";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/scanner", label: "Wallet scanner", icon: ScanSearch },
  { href: "/reports", label: "Reports", icon: FileText },
  { href: "/quantum", label: "Quantum lab", icon: FlaskConical },
  { href: "/learn", label: "Research & learn", icon: BookOpen },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="flex h-full w-60 shrink-0 flex-col border-r border-border bg-[#0D1117]">
      <div className="flex items-center gap-2 border-b border-border px-5 py-5">
        <div className="flex h-9 w-9 items-center justify-center rounded-md border border-cyan/30 bg-cyan-muted">
          <Shield className="h-5 w-5 text-cyan" aria-hidden />
        </div>
        <div>
          <p className="text-sm font-semibold tracking-wide text-foreground">Q-Guard</p>
          <p className="text-[11px] text-muted-foreground">Quantum threat scanner</p>
        </div>
      </div>
      <nav className="flex flex-1 flex-col gap-1 p-3" aria-label="Primary">
        {nav.map((item) => {
          const active =
            item.href === "/"
              ? pathname === "/"
              : pathname === item.href || pathname.startsWith(`${item.href}/`);
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
                active
                  ? "bg-cyan-muted text-cyan"
                  : "text-muted-foreground hover:bg-muted/50 hover:text-foreground",
              )}
            >
              <Icon className="h-4 w-4" aria-hidden />
              {item.label}
            </Link>
          );
        })}
      </nav>
      <div className="border-t border-border p-4 text-[11px] leading-relaxed text-muted-foreground">
        Public addresses only. No wallet connect. No private keys.
      </div>
    </aside>
  );
}
