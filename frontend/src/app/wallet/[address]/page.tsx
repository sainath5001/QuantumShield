import { Suspense } from "react";
import WalletAnalysisClient from "./page-client";
import { Skeleton } from "@/components/ui/skeleton";

export default function WalletAnalysisPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-3 p-6">
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-48 w-full" />
        </div>
      }
    >
      <WalletAnalysisClient />
    </Suspense>
  );
}
