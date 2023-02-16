import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";

export function filterEntriesOnReference(modelName, references) {
  return function (entry: V1CatalogEntry) {
    return references?.some((ref) => {
      return (
        ref.reference === entry.name ||
        entry?.children?.includes(modelName.toLowerCase())
      );
    });
  };
}

export function combineEntryWithReference(modelName, references) {
  return function (entry: V1CatalogEntry) {
    // get the reference that matches this entry
    return [
      entry,
      references.find(
        (ref) =>
          ref.reference === entry.name ||
          (entry?.embedded &&
            entry?.children?.includes(modelName.toLowerCase()))
      ),
    ];
  };
}

export function getMatchingReferencesAndEntries(
  modelName,
  references,
  entries: V1CatalogEntry[]
) {
  return entries
    ?.filter(filterEntriesOnReference(modelName, references))
    ?.map(combineEntryWithReference(modelName, references));
}
