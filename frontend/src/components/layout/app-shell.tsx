import { MobileNav } from "@/components/layout/mobile-nav";
import { Sidebar } from "@/components/layout/sidebar";

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen bg-background text-foreground">
      <div className="hidden md:block">
        <Sidebar />
      </div>
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="space-y-2 border-b border-border px-4 py-3 md:hidden">
          <p className="text-sm font-semibold text-cyan">Q-Guard</p>
          <MobileNav />
        </div>
        <main className="flex-1">{children}</main>
      </div>
    </div>
  );
}
