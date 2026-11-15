import { describe, expect, it } from "vitest";
import { ethereumAddressSchema, isValidEthereumAddress, scanFormSchema } from "./address";

describe("ethereumAddressSchema", () => {
  it("accepts valid checksummed and lowercase addresses", () => {
    expect(isValidEthereumAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")).toBe(true);
    expect(isValidEthereumAddress("0x742d35cc6634c0532925a3b844bc9e7595f0beb0")).toBe(true);
  });

  it("rejects missing 0x, wrong length, and non-hex", () => {
    expect(isValidEthereumAddress("742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")).toBe(false);
    expect(isValidEthereumAddress("0x123")).toBe(false);
    expect(isValidEthereumAddress("0xGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG")).toBe(false);
  });

  it("powers scan form schema", () => {
    expect(scanFormSchema.safeParse({ address: "not-an-address" }).success).toBe(false);
    expect(
      ethereumAddressSchema.parse("0x0000000000000000000000000000000000000001"),
    ).toBe("0x0000000000000000000000000000000000000001");
  });
});
