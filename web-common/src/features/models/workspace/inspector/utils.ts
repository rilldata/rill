import type { Reference } from "@rilldata/web-common/features/models/utils/get-table-references";
import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";

export function filterEntriesOnReference(
  modelName: string,
  references: Reference[]
) {
  return function (entry: V1CatalogEntry) {
    return references?.some((ref) => {
      return (
        ref.reference === entry.name ||
        entry?.children?.includes(modelName.toLowerCase())
      );
    });
  };
}

export function combineEntryWithReference(
  modelName: string,
  references: Reference[]
) {
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
  modelName: string,
  references: Reference[],
  entries: V1CatalogEntry[]
) {
  return entries
    ?.filter(filterEntriesOnReference(modelName, references))
    ?.map(combineEntryWithReference(modelName, references));
}
