import type { TimeSeriesPoint } from "./types";

/**
 * Clamp a fractional index to the nearest valid array index.
 */
export function snapIndex(idx: number, length: number): number {
  return Math.max(0, Math.min(length - 1, Math.round(idx)));
}

/**
 * Find the data index whose timestamp is closest to the given millisecond timestamp.
 */
export function dateToIndex(
  data: TimeSeriesPoint[],
  ms: number,
): number | null {
  if (data.length === 0) return null;
  let best = 0;
  let bestDist = Infinity;
  for (let i = 0; i < data.length; i++) {
    const dist = Math.abs(data[i].ts.toMillis() - ms);
    if (dist < bestDist) {
      bestDist = dist;
      best = i;
    }
  }
  return best;
}

export interface BarSlotGeometry {
  slotWidth: number;
  gap: number;
  bandWidth: number;
  barGap: number;
  singleBarWidth: number;
}

/**
 * Compute the bar slot geometry for grouped bar charts.
 */
export function computeBarSlotGeometry(
  plotWidth: number,
  visibleCount: number,
  barCount: number,
): BarSlotGeometry {
  const slotWidth = plotWidth / Math.max(1, visibleCount);
  const gap = slotWidth * 0.2;
  const bandWidth = slotWidth - gap;
  const barGap = barCount > 1 ? 2 : 0;
  const totalGaps = barGap * (barCount - 1);
  const singleBarWidth = (bandWidth - totalGaps) / barCount;
  return { slotWidth, gap, bandWidth, barGap, singleBarWidth };
}

/**
 * Compute the x position of a bar center within a slot.
 */
export function barCenterX(
  slotCenterX: number,
  bandWidth: number,
  singleBarWidth: number,
  barGap: number,
  barIndex: number,
): number {
  return (
    slotCenterX -
    bandWidth / 2 +
    barIndex * (singleBarWidth + barGap) +
    singleBarWidth / 2
  );
}
