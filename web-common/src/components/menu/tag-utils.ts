// Pure helpers for deriving tag metadata and applying tag-driven visibility
// changes to a flat list of taggable items (dimensions or measures).
//
// The dimensions and measures dropdown is the only caller today; if other
// surfaces need tag-aware behavior they should consume this module directly
// rather than re-deriving the same logic.

export type Taggable = {
  name?: string;
  displayName?: string;
  tags?: string[];
};

export type DimensionTag = {
  name: string;
  displayName: string;
  totalCount: number;
};

export type TagVisibilityState = {
  tagName: string;
  visibleCount: number;
  totalCount: number;
  state: "none" | "partial" | "all";
};

// Walks items in spec order and collects unique tags. Tag strings are trimmed
// to normalize accidental whitespace from YAML authoring. Tags remain
// case-sensitive: "Geography" and "geography" are distinct.
export function deriveTags(items: Taggable[]): DimensionTag[] {
  const seen = new Map<string, number>();
  for (const item of items) {
    if (!item.tags) continue;
    for (const raw of item.tags) {
      const tag = typeof raw === "string" ? raw.trim() : "";
      if (!tag) continue;
      seen.set(tag, (seen.get(tag) ?? 0) + 1);
    }
  }
  return Array.from(seen, ([name, total]) => ({
    name,
    displayName: name,
    totalCount: total,
  }));
}

export function itemHasTag(item: Taggable, tagName: string): boolean {
  if (!item.tags) return false;
  for (const t of item.tags) {
    if (typeof t === "string" && t.trim() === tagName) return true;
  }
  return false;
}

export function namesInTag(items: Taggable[], tagName: string): string[] {
  const result: string[] = [];
  for (const item of items) {
    if (!item.name || !itemHasTag(item, tagName)) continue;
    result.push(item.name);
  }
  return result;
}

export function computeTagVisibility(
  items: Taggable[],
  visibleNames: Iterable<string>,
  tagName: string,
): TagVisibilityState {
  const visible =
    visibleNames instanceof Set ? visibleNames : new Set(visibleNames);
  let total = 0;
  let visibleCount = 0;
  for (const item of items) {
    if (!itemHasTag(item, tagName)) continue;
    total += 1;
    if (item.name && visible.has(item.name)) visibleCount += 1;
  }
  return {
    tagName,
    visibleCount,
    totalCount: total,
    state:
      visibleCount === 0 ? "none" : visibleCount === total ? "all" : "partial",
  };
}

// Union of currently visible names with all items in tagName, ordered by spec.
export function applyShowAllInTag(
  visibleNames: string[],
  items: Taggable[],
  tagName: string,
): string[] {
  const union = new Set([...visibleNames, ...namesInTag(items, tagName)]);
  return orderBySpec(union, items);
}

// Currently visible minus items in tagName. Clamped to keep one item visible
// to match the existing "at least one dimension/measure must remain shown"
// invariant in setDimensionVisibility / setMeasureVisibility.
export function applyHideAllInTag(
  visibleNames: string[],
  items: Taggable[],
  tagName: string,
): string[] {
  const remove = new Set(namesInTag(items, tagName));
  const remaining = visibleNames.filter((n) => !remove.has(n));
  return clampMinOne(
    orderBySpec(new Set(remaining), items),
    visibleNames,
    items,
  );
}

// Sets visible names to exactly the items in tagName. Clamped to keep one
// item visible.
export function applyOnlyShowTag(
  visibleNames: string[],
  items: Taggable[],
  tagName: string,
): string[] {
  const inTag = namesInTag(items, tagName);
  return clampMinOne(orderBySpec(new Set(inTag), items), visibleNames, items);
}

function orderBySpec(allowed: Set<string>, items: Taggable[]): string[] {
  const result: string[] = [];
  for (const item of items) {
    if (item.name && allowed.has(item.name)) result.push(item.name);
  }
  return result;
}

function clampMinOne(
  next: string[],
  current: string[],
  items: Taggable[],
): string[] {
  if (next.length > 0) return next;
  const fallback = current[0] ?? items[0]?.name;
  return fallback ? [fallback] : [];
}
