// Live Connect tiers shared between the modal and tests
export const LIVE_CONNECT_TIERS = [
  { slots: 4, instance: "8GiB / 2vCPU", rillBill: 99 },
  { slots: 6, instance: "12GiB / 3vCPU", rillBill: 130 },
  { slots: 8, instance: "16GiB / 4vCPU", rillBill: 175 },
  { slots: 16, instance: "32GiB / 8vCPU", rillBill: 350 },
  { slots: 32, instance: "64GiB / 16vCPU", rillBill: 700 },
  { slots: 60, instance: "120GiB / 30vCPU", rillBill: 1300 },
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
