// Legacy rate for existing Team plan customers
export const SLOT_RATE_PER_HR = 0.06;
export const HOURS_PER_MONTH = 730;

// New pricing rates (PRD v10)
export const MANAGED_SLOT_RATE_PER_HR = 0.15;
export const CLUSTER_SLOT_RATE_PER_HR = 0.06;
export const RILL_SLOT_RATE_PER_HR = 0.15;
export const STORAGE_RATE_PER_GB_PER_MONTH = 1.0;
export const INCLUDED_STORAGE_GB = 1;

export interface SlotTier {
  slots: number;
  instance: string;
  rillBill: number;
}

function tier(slots: number, rate = SLOT_RATE_PER_HR): SlotTier {
  return {
    slots,
    instance: `${slots * 4}GiB / ${slots}vCPU`,
    rillBill: Math.round(slots * rate * HOURS_PER_MONTH),
  };
}

// Popular slot values shown by default
export const POPULAR_SLOTS = [2, 3, 4, 8, 16, 30];

// All available slot values including intermediate sizes
export const ALL_SLOTS = [2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16, 20, 24, 28, 30];

// Legacy tiers at $0.06/slot/hr (for existing Team plan customers)
export const POPULAR_LIVE_CONNECT_TIERS: SlotTier[] = POPULAR_SLOTS.map((s) =>
  tier(s),
);
export const LIVE_CONNECT_TIERS: SlotTier[] = ALL_SLOTS.map((s) => tier(s));

// New pricing tiers: Cluster Slots at $0.06/hr (auto-calculated, read-only for Live Connect)
export const CLUSTER_SLOT_TIERS: SlotTier[] = ALL_SLOTS.map((s) =>
  tier(s, CLUSTER_SLOT_RATE_PER_HR),
);

// New pricing tiers: Rill Slots at $0.15/hr (user-controlled for Live Connect)
export const RILL_SLOT_TIERS: SlotTier[] = ALL_SLOTS.map((s) =>
  tier(s, RILL_SLOT_RATE_PER_HR),
);

// Managed mode tiers at $0.15/slot/hr
export const MANAGED_SLOT_TIERS: SlotTier[] = ALL_SLOTS.map((s) =>
  tier(s, MANAGED_SLOT_RATE_PER_HR),
);

/**
 * Given detected cluster memory (GB per replica), return the matching tier's slot count.
 * Picks the tier whose memory (slots * 4 GB) is closest to the detected value.
 */
export function detectTierSlots(
  detectedMemoryGb: number | undefined,
): number | undefined {
  if (!detectedMemoryGb) return undefined;
  const match = LIVE_CONNECT_TIERS.reduce((best, tier) => {
    const tierMemory = tier.slots * 4;
    const bestMemory = best.slots * 4;
    return Math.abs(tierMemory - detectedMemoryGb) <
      Math.abs(bestMemory - detectedMemoryGb)
      ? tier
      : best;
  });
  return match.slots;
}
