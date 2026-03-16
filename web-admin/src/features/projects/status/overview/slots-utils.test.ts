import { describe, it, expect } from "vitest";
import {
  detectTierSlots,
  LIVE_CONNECT_TIERS,
  POPULAR_LIVE_CONNECT_TIERS,
} from "./slots-utils";

describe("detectTierSlots", () => {
  it("returns undefined for undefined input", () => {
    expect(detectTierSlots(undefined)).toBeUndefined();
  });

  it("returns undefined for 0", () => {
    expect(detectTierSlots(0)).toBeUndefined();
  });

  it("matches exact tier memories", () => {
    // Each tier's memory = slots * 4
    expect(detectTierSlots(8)).toBe(2); // 2 slots * 4 = 8 GB
    expect(detectTierSlots(12)).toBe(3); // 3 slots * 4 = 12 GB
    expect(detectTierSlots(16)).toBe(4);
    expect(detectTierSlots(32)).toBe(8);
    expect(detectTierSlots(64)).toBe(16);
    expect(detectTierSlots(120)).toBe(30);
  });

  it("rounds to the closest tier", () => {
    // Tier memories (slots * 4): 8, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, ...
    // 10 GB: |10-8|=2, |10-12|=2 → tie, keeps first (2 slots)
    expect(detectTierSlots(10)).toBe(2);
    // 11 GB: |11-8|=3, |11-12|=1 → 3 slots
    expect(detectTierSlots(11)).toBe(3);
    // 14 GB: |14-12|=2, |14-16|=2 → tie, keeps 3 slots
    expect(detectTierSlots(14)).toBe(3);
    // 15 GB: |15-12|=3, |15-16|=1 → 4 slots
    expect(detectTierSlots(15)).toBe(4);
  });

  it("handles memory below smallest tier", () => {
    expect(detectTierSlots(4)).toBe(2);
    expect(detectTierSlots(1)).toBe(2);
  });

  it("handles memory above largest tier", () => {
    expect(detectTierSlots(200)).toBe(30);
    expect(detectTierSlots(500)).toBe(30);
  });

  it("all tiers list has expected entries", () => {
    expect(LIVE_CONNECT_TIERS).toHaveLength(15);
  });

  it("popular tiers list has expected entries", () => {
    expect(POPULAR_LIVE_CONNECT_TIERS).toHaveLength(6);
  });
});
