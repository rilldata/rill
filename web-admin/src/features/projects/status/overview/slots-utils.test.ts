import { describe, it, expect } from "vitest";
import { detectTierSlots, LIVE_CONNECT_TIERS } from "./slots-utils";

describe("detectTierSlots", () => {
  it("returns undefined for undefined input", () => {
    expect(detectTierSlots(undefined)).toBeUndefined();
  });

  it("returns undefined for 0", () => {
    expect(detectTierSlots(0)).toBeUndefined();
  });

  it("matches exact tier memories", () => {
    // Each tier's memory = slots * 2
    expect(detectTierSlots(8)).toBe(4); // 4 slots * 2 = 8 GB
    expect(detectTierSlots(12)).toBe(6); // 6 slots * 2 = 12 GB
    expect(detectTierSlots(16)).toBe(8);
    expect(detectTierSlots(32)).toBe(16);
    expect(detectTierSlots(64)).toBe(32);
    expect(detectTierSlots(120)).toBe(60);
  });

  it("rounds to the closest tier", () => {
    // Tier memory = slots * 2: 8, 12, 16, 32, 64, 120
    // 10 GB: |10-8|=2, |10-12|=2 → tie, reduce keeps first (4 slots)
    expect(detectTierSlots(10)).toBe(4);
    // 11 GB: |11-8|=3, |11-12|=1 → closer to 12 (6 slots)
    expect(detectTierSlots(11)).toBe(6);
    // 14 GB: |14-12|=2, |14-16|=2 → tie, keeps 6 slots
    expect(detectTierSlots(14)).toBe(6);
    // 15 GB: |15-12|=3, |15-16|=1 → closer to 16 (8 slots)
    expect(detectTierSlots(15)).toBe(8);
    // 24 GB: |24-16|=8, |24-32|=8 → tie, keeps 8 slots
    expect(detectTierSlots(24)).toBe(8);
  });

  it("handles memory below smallest tier", () => {
    expect(detectTierSlots(4)).toBe(4);
    expect(detectTierSlots(1)).toBe(4);
  });

  it("handles memory above largest tier", () => {
    expect(detectTierSlots(200)).toBe(60);
    expect(detectTierSlots(500)).toBe(60);
  });

  it("tier list has expected number of entries", () => {
    expect(LIVE_CONNECT_TIERS).toHaveLength(6);
  });
});
