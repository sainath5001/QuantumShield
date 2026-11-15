import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { FindingBadge, RiskBadge } from "./risk-badge";

describe("RiskBadge", () => {
  it("renders elevated label", () => {
    render(<RiskBadge classification="ELEVATED_EXPOSURE" />);
    expect(screen.getByText(/Elevated exposure/i)).toBeInTheDocument();
  });

  it("renders empty assessment state", () => {
    render(<RiskBadge classification={null} />);
    expect(screen.getByText(/No assessment/i)).toBeInTheDocument();
  });
});

describe("FindingBadge", () => {
  it("labels confirmed exposure", () => {
    render(<FindingBadge finding="public_key_exposed" />);
    expect(screen.getByText(/Public key exposed/i)).toBeInTheDocument();
  });
});
