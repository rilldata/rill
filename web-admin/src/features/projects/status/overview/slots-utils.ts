const SLOT_RATE_PER_HR = 0.03;
const HOURS_PER_MONTH = 730;

function tier(slots: number) {
  return {
    slots,
    instance: `${slots * 2}GiB / ${Math.max(1, slots / 2)}vCPU`,
    rillBill: Math.round(slots * SLOT_RATE_PER_HR * HOURS_PER_MONTH),
  };
}

// Popular tiers shown by default
export const POPULAR_LIVE_CONNECT_TIERS = [
  tier(4),
  tier(6),
  tier(8),
  tier(16),
  tier(32),
  tier(60),
];

// All available tiers including intermediate sizes
export const LIVE_CONNECT_TIERS = [
  tier(4),
  tier(6),
  tier(8),
  tier(10),
  tier(12),
  tier(14),
  tier(16),
  tier(20),
  tier(24),
  tier(28),
  tier(32),
  tier(36),
  tier(40),
  tier(44),
  tier(48),
  tier(52),
  tier(56),
  tier(60),
];

/**
 * Given detected cluster memory (GB per replica), return the matching tier's slot count.
 * Picks the tier whose memory (slots * 2 GB) is closest to the detected value.
 */
export function detectTierSlots(
  detectedMemoryGb: number | undefined,
): number | undefined {
  if (!detectedMemoryGb) return undefined;
  const match = LIVE_CONNECT_TIERS.reduce((best, tier) => {
    const tierMemory = tier.slots * 2;
    const bestMemory = best.slots * 2;
    return Math.abs(tierMemory - detectedMemoryGb) <
      Math.abs(bestMemory - detectedMemoryGb)
      ? tier
      : best;
  });
  return match.slots;
}
