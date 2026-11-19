import type { PivotChipData } from "./types";

/**
 * Drops the chips that break the pivot: an empty id, an id in `excludeIds`
 * and any repeat of an id seen earlier in the list. The first occurrence
 * wins, so the zone keeps its order. The drop zones render chips keyed by
 * id, so a repeated id is a keyed-each error that takes the whole view down.
 */
export function dedupePivotChips(
  chips: PivotChipData[],
  excludeIds: Iterable<string> = [],
): PivotChipData[] {
  const seen = new Set(excludeIds);
  return chips.filter((chip) => {
    if (!chip.id || seen.has(chip.id)) return false;
    seen.add(chip.id);
    return true;
  });
}

/**
 * Returns `items` with `chip` placed at `index`, removing any copy of the
 * chip that is already in the list. Dropping a chip back into the zone it
 * came from, or into a zone whose list was reset from the store while the
 * drag was in flight, then moves the chip instead of adding a second copy.
 * `index` counts positions in `items` as given, including the removed copy.
 */
export function insertPivotChipAt(
  items: PivotChipData[],
  chip: PivotChipData,
  index: number,
): PivotChipData[] {
  const removedBefore = items
    .slice(0, index)
    .filter((item) => item.id === chip.id).length;
  const next = items.filter((item) => item.id !== chip.id);
  next.splice(index - removedBefore, 0, chip);
  return next;
}
