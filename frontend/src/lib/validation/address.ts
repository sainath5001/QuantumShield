import { z } from "zod";

/** Ethereum address: 0x + 40 hex chars (checksum optional). */
export const ethereumAddressSchema = z
  .string()
  .trim()
  .regex(/^0x[a-fA-F0-9]{40}$/, "Enter a valid Ethereum address (0x + 40 hex characters).");

export type EthereumAddress = z.infer<typeof ethereumAddressSchema>;

export function isValidEthereumAddress(value: string): boolean {
  return ethereumAddressSchema.safeParse(value).success;
}

export const scanFormSchema = z.object({
  address: ethereumAddressSchema,
});

export type ScanFormValues = z.infer<typeof scanFormSchema>;
