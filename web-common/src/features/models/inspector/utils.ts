import type { Reference } from "@rilldata/web-common/features/models/utils/get-table-references";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

function resourceHasReference(resource: V1Resource, name: string) {
  return (
    resource.meta.refs.findIndex(
      (resRef) => resRef.name.toLowerCase() === name.toLowerCase(),
    ) !== -1
  );
}

function filterEntriesOnReference(
  modelName: string,
  references: Array<Reference>,
) {
  return function (resource: V1Resource) {
    return references?.some((ref) => {
      return (
        ref.reference === resource.meta.name.name ||
        resourceHasReference(resource, modelName)
      );
    });
  };
}

function combineEntryWithReference(references: Array<Reference>) {
  return function (resource: V1Resource) {
    // get the reference that matches this entry
    return [
      resource,
      references.find((ref) => ref.reference === resource.meta.name.name),
    ] as [V1Resource, Reference];
  };
}

export function getMatchingReferencesAndEntries(
  modelName: string,
  references: Array<Reference>,
  resources: Array<V1Resource>,
) {
  return resources
    ?.filter(filterEntriesOnReference(modelName, references))
    ?.map(combineEntryWithReference(references));
}
