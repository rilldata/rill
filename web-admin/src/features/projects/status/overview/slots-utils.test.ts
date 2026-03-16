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
    // Each tier's memory = slots * 2
    expect(detectTierSlots(8)).toBe(4);
    expect(detectTierSlots(12)).toBe(6);
    expect(detectTierSlots(16)).toBe(8);
    expect(detectTierSlots(20)).toBe(10);
    expect(detectTierSlots(24)).toBe(12);
    expect(detectTierSlots(28)).toBe(14);
    expect(detectTierSlots(32)).toBe(16);
    expect(detectTierSlots(64)).toBe(32);
    expect(detectTierSlots(120)).toBe(60);
  });

  it("rounds to the closest tier", () => {
    // Tier memories: 8, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, ...
    // 11 GB: |11-8|=3, |11-12|=1 → 6 slots
    expect(detectTierSlots(11)).toBe(6);
    // 15 GB: |15-12|=3, |15-16|=1 → 8 slots
    expect(detectTierSlots(15)).toBe(8);
    // 18 GB: |18-16|=2, |18-20|=2 → tie, keeps first (8 slots)
    expect(detectTierSlots(18)).toBe(8);
    // 19 GB: |19-16|=3, |19-20|=1 → 10 slots
    expect(detectTierSlots(19)).toBe(10);
  });

  it("handles memory below smallest tier", () => {
    expect(detectTierSlots(4)).toBe(4);
    expect(detectTierSlots(1)).toBe(4);
  });

  it("handles memory above largest tier", () => {
    expect(detectTierSlots(200)).toBe(60);
    expect(detectTierSlots(500)).toBe(60);
  });

  it("all tiers list has expected entries", () => {
    expect(LIVE_CONNECT_TIERS).toHaveLength(18);
  });

  it("popular tiers list has expected entries", () => {
    expect(POPULAR_LIVE_CONNECT_TIERS).toHaveLength(6);
  });
});
