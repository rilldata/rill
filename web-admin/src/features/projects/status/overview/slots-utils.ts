export const SLOT_RATE_PER_HR = 0.06;
export const HOURS_PER_MONTH = 730;

export interface SlotTier {
  slots: number;
  instance: string;
  rillBill: number;
}

function tier(slots: number): SlotTier {
  return {
    slots,
    instance: `${slots * 4}GiB / ${slots}vCPU`,
    rillBill: Math.round(slots * SLOT_RATE_PER_HR * HOURS_PER_MONTH),
  };
}

// Popular slot values shown by default
export const POPULAR_SLOTS = [2, 3, 4, 8, 16, 30];

// All available slot values including intermediate sizes
export const ALL_SLOTS = [2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16, 20, 24, 28, 30];

// Popular tiers shown by default
export const POPULAR_LIVE_CONNECT_TIERS: SlotTier[] = POPULAR_SLOTS.map(tier);

// All available tiers including intermediate sizes
export const LIVE_CONNECT_TIERS: SlotTier[] = ALL_SLOTS.map(tier);

/**
 * Given detected cluster memory (GB per replica), return the matching tier's slot count.
 * Picks the tier whose memory (slots * 2 GB) is closest to the detected value.
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
