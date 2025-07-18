import type { PivotChipData } from "@rilldata/web-common/features/dashboards/pivot/types";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  getLargestGrain,
  getNextSmallerGrain,
  isAvailableTimeGrain,
} from "@rilldata/web-common/lib/time/grains";
import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { PivotChipType } from "./types";

/**
 * Get the next smaller grain from the closest time pill to the left
 */
export function getNextGrainFromPosition(
  dropPosition: number,
  timeChips: PivotChipData[],
  availableGrains: AvailableTimeGrain[],
): AvailableTimeGrain | undefined {
  // Find time chips to the left of the drop position
  const timeChipsToLeft = timeChips.slice(0, dropPosition).reverse();

  if (timeChipsToLeft.length === 0) {
    return getLargestGrain(availableGrains);
  }

  // Get the closest time chip's grain
  const closestGrain = timeChipsToLeft[0].id as V1TimeGrain;

  // Find the next smaller grain
  if (isAvailableTimeGrain(closestGrain)) {
    return getNextSmallerGrain(closestGrain, availableGrains);
  }

  return getLargestGrain(availableGrains);
}

/**
 * Create a time chip with a specific grain
 */
export function createTimeChipWithGrain(grain: V1TimeGrain): PivotChipData {
  const grainConfig = TIME_GRAIN[grain];
  return {
    id: grain,
    title: grainConfig?.label || grain,
    type: PivotChipType.Time,
  };
}

export function isNewTimeChip(pivotChip: PivotChipData) {
  if (pivotChip.type !== PivotChipType.Time) return false;
  const allGrains = Object.keys(TIME_GRAIN);
  return !allGrains.includes(pivotChip.id);
}

/**
 * Handle time chip transformation when dropping
 */
export function handleTimeChipDrop(
  chip: PivotChipData,
  dropPosition: number,
  existingTimeChips: PivotChipData[],
  availableGrains: AvailableTimeGrain[],
): PivotChipData {
  let selectedGrain: V1TimeGrain | undefined;

  if (existingTimeChips.length === 0) {
    // No time chips in the drop zone, use the largest available grain
    selectedGrain = getLargestGrain(availableGrains);
  } else {
    // Get the appropriate grain based on position relative to existing time chips
    selectedGrain = getNextGrainFromPosition(
      dropPosition,
      existingTimeChips,
      availableGrains,
    );
  }

  if (selectedGrain) {
    return createTimeChipWithGrain(selectedGrain);
  }

  return chip;
}

/**
 * Handle time chip click for adding to pivot
 */
export function handleTimeChipClick(
  chip: PivotChipData,
  availableGrains: AvailableTimeGrain[],
): PivotChipData {
  const selectedGrain = getLargestGrain(availableGrains);
  if (selectedGrain) {
    return createTimeChipWithGrain(selectedGrain);
  }

  return chip;
}

/**
 * Update a time chip with a new grain
 */
export function updateTimeChipGrain(
  items: PivotChipData[],
  targetChip: PivotChipData,
  newGrain: V1TimeGrain,
): PivotChipData[] {
  return items.map((chip) =>
    chip.id === targetChip.id ? createTimeChipWithGrain(newGrain) : chip,
  );
}
