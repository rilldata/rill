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
  totalCount: number;
};

export type TagVisibilityState = {
  tagName: string;
  visibleCount: number;
  totalCount: number;
  state: "none" | "partial" | "all";
};

// Cached lookup table built once from a list of items. Operations below walk
// the (typically small) per-tag bucket instead of re-scanning every item.
//
// Tag strings are trimmed to normalize accidental whitespace from YAML
// authoring. Tags remain case-sensitive: "Geography" and "geography" are
// distinct.
export type TagIndex = {
  tags: DimensionTag[];
  itemsByTag: Map<string, Taggable[]>;
  allItems: Taggable[];
};

// Tags listed in tagOrder (the metrics view's `tag_order`) come first, in that
// order. The rest follow in the order they first appear on an item.
export function buildTagIndex(
  items: Taggable[],
  tagOrder?: string[],
): TagIndex {
  const itemsByTag = new Map<string, Taggable[]>();
  for (const item of items) {
    if (!item.tags) continue;
    for (const raw of item.tags) {
      const tag = typeof raw === "string" ? raw.trim() : "";
      if (!tag) continue;
      let bucket = itemsByTag.get(tag);
      if (!bucket) {
        bucket = [];
        itemsByTag.set(tag, bucket);
      }
      bucket.push(item);
    }
  }
  const tags: DimensionTag[] = [];
  for (const [name, bucket] of itemsByTag) {
    tags.push({ name, totalCount: bucket.length });
  }
  // The REST API can send an unset list as null.
  const order = tagOrder ?? [];
  if (order.length > 0) {
    const rank = new Map(order.map((tag, i) => [tag.trim(), i]));
    // Unlisted tags share the last rank; the stable sort keeps their order.
    tags.sort(
      (a, b) =>
        (rank.get(a.name) ?? order.length) - (rank.get(b.name) ?? order.length),
    );
  }
  return { tags, itemsByTag, allItems: items };
}

export function itemsInTag(index: TagIndex, tagName: string): Taggable[] {
  return index.itemsByTag.get(tagName) ?? [];
}

export function namesInTag(index: TagIndex, tagName: string): string[] {
  const result: string[] = [];
  for (const item of itemsInTag(index, tagName)) {
    if (item.name) result.push(item.name);
  }
  return result;
}

export function computeTagVisibility(
  index: TagIndex,
  visibleNames: Iterable<string>,
  tagName: string,
): TagVisibilityState {
  const visible =
    visibleNames instanceof Set ? visibleNames : new Set(visibleNames);
  const bucket = itemsInTag(index, tagName);
  let visibleCount = 0;
  for (const item of bucket) {
    if (item.name && visible.has(item.name)) visibleCount += 1;
  }
  const total = bucket.length;
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
  index: TagIndex,
  tagName: string,
): string[] {
  const union = new Set([...visibleNames, ...namesInTag(index, tagName)]);
  return orderBySpec(union, index.allItems);
}

// Currently visible minus items in tagName. Clamped to keep one item visible
// to match the existing "at least one dimension/measure must remain shown"
// invariant in setDimensionVisibility / setMeasureVisibility.
export function applyHideAllInTag(
  visibleNames: string[],
  index: TagIndex,
  tagName: string,
): string[] {
  const remove = new Set(namesInTag(index, tagName));
  const remaining = visibleNames.filter((n) => !remove.has(n));
  return clampMinOne(
    orderBySpec(new Set(remaining), index.allItems),
    visibleNames,
    index.allItems,
  );
}

// Sets visible names to exactly the items in tagName. Clamped to keep one
// item visible.
export function applyOnlyShowTag(
  visibleNames: string[],
  index: TagIndex,
  tagName: string,
): string[] {
  const inTag = namesInTag(index, tagName);
  return clampMinOne(
    orderBySpec(new Set(inTag), index.allItems),
    visibleNames,
    index.allItems,
  );
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
