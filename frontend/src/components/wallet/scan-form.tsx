"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { ScanSearch } from "lucide-react";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { scanFormSchema, type ScanFormValues } from "@/lib/validation/address";

export function ScanForm({
  onSubmit,
  loading,
  defaultAddress = "",
}: {
  onSubmit: (values: ScanFormValues) => void | Promise<void>;
  loading?: boolean;
  defaultAddress?: string;
}) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ScanFormValues>({
    resolver: zodResolver(scanFormSchema),
    defaultValues: { address: defaultAddress },
  });

  return (
    <form className="space-y-4" onSubmit={handleSubmit(onSubmit)} noValidate>
      <div className="space-y-2">
        <Label htmlFor="address">Ethereum address</Label>
        <Input
          id="address"
          placeholder="0x…"
          autoComplete="off"
          spellCheck={false}
          aria-invalid={Boolean(errors.address)}
          aria-describedby={errors.address ? "address-error" : "address-help"}
          {...register("address")}
        />
        <p id="address-help" className="text-xs text-muted-foreground">
          Public address only. Q-Guard never asks for private keys, seed phrases, or wallet connection.
        </p>
        {errors.address ? (
          <p id="address-error" className="text-xs text-destructive" role="alert">
            {errors.address.message}
          </p>
        ) : null}
      </div>
      <Button type="submit" variant="cyan" disabled={loading} className="w-full sm:w-auto">
        <ScanSearch className="h-4 w-4" aria-hidden />
        {loading ? "Starting scan…" : "Scan wallet"}
      </Button>
    </form>
  );
}
