import { describe, it, expect } from "vitest";
import {
  SLOT_TIERS,
  POPULAR_SLOTS,
  ALL_SLOTS,
  DEFAULT_MANAGED_SLOTS,
  DEFAULT_SELF_MANAGED_SLOTS,
} from "./slots-utils";

describe("slots-utils", () => {
  it("all tiers list has expected entries", () => {
    expect(SLOT_TIERS).toHaveLength(ALL_SLOTS.length);
  });

  it("popular slots list has expected entries", () => {
    expect(POPULAR_SLOTS).toHaveLength(7);
  });

  it("managed default is 2 slots", () => {
    expect(DEFAULT_MANAGED_SLOTS).toBe(2);
  });

  it("self-managed default is 4 slots", () => {
    expect(DEFAULT_SELF_MANAGED_SLOTS).toBe(4);
  });

  it("all slot values are at least 1", () => {
    for (const s of ALL_SLOTS) {
      expect(s).toBeGreaterThanOrEqual(1);
    }
  });

  it("tiers have correct bill calculations", () => {
    const tier = SLOT_TIERS[0]; // 1 slot
    expect(tier.slots).toBe(1);
    expect(tier.rillBill).toBe(Math.round(1 * 0.15 * 730));
  });
});
