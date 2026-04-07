import { describe, it, expect } from "vitest";
import {
  SLOT_TIERS,
  POPULAR_SLOTS,
  ALL_SLOTS,
  DEFAULT_MANAGED_SLOTS,
} from "./slots-utils";

describe("slots-utils", () => {
  it("all tiers list has expected entries", () => {
    expect(SLOT_TIERS).toHaveLength(ALL_SLOTS.length);
  });

  it("popular slots list has expected entries", () => {
    expect(POPULAR_SLOTS).toHaveLength(6);
  });

  it("default minimum is 2 slots", () => {
    expect(DEFAULT_MANAGED_SLOTS).toBe(2);
  });

  it("all slot values are at least minimum", () => {
    for (const s of ALL_SLOTS) {
      expect(s).toBeGreaterThanOrEqual(DEFAULT_MANAGED_SLOTS);
    }
  });

  it("tiers have correct bill calculations", () => {
    const tier = SLOT_TIERS[0]; // 2 slots
    expect(tier.slots).toBe(2);
    expect(tier.rillBill).toBe(Math.round(2 * 0.15 * 730));
  });
});
