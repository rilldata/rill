export const SLOT_RATE_PER_HR = 0.15;
export const HOURS_PER_MONTH = 730;

// Minimum slots for any deployment
export const DEFAULT_MANAGED_SLOTS = 2;

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
export const POPULAR_SLOTS = [1, 2, 3, 4, 8, 16, 30];

// All available slot values including intermediate sizes
export const ALL_SLOTS = [
  1, 2, 3, 4, 5, 6, 7, 8, 10, 12, 14, 16, 20, 24, 28, 30,
];

export const SLOT_TIERS: SlotTier[] = ALL_SLOTS.map((s) => tier(s));
