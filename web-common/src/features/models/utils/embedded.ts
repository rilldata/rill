import type { Reference } from "@rilldata/web-common/features/models/utils/get-table-references";
import { humanReadableErrorMessage } from "@rilldata/web-common/features/sources/add-source/errors";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import type {
  V1CatalogEntry,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";

export function filterKnownEmbeddedSources(
  embeddedRefs: Array<Reference>,
  embeddedSourceCatalogs: Map<string, V1CatalogEntry>
): Array<string> {
  const unknownEmbeddedSources = new Array<string>();
  for (const embeddedRef of embeddedRefs) {
    const cleanedRef = embeddedRef.reference.slice(
      1,
      embeddedRef.reference.length - 1
    );
    const ref = cleanedRef.toLowerCase();
    if (embeddedSourceCatalogs.has(ref)) continue;
    unknownEmbeddedSources.push(cleanedRef);
  }
  return unknownEmbeddedSources;
}

export function embeddedSourcesError(
  errors: Array<V1ReconcileError>,
  embeddedSources: Array<Reference>
) {
  const embeddedSourcesMap = getMapFromArray(embeddedSources, (entity) =>
    entity.reference.slice(1, entity.reference.length - 1)
  );
  const embeddedSourceErrors = new Array<string>();
  for (const reconcileError of errors) {
    if (!embeddedSourcesMap.has(reconcileError.filePath.toLowerCase())) {
      continue;
    }
    embeddedSourceErrors.push(
      `${reconcileError.filePath} - ${humanReadableErrorMessage(
        reconcileError.filePath.replace(/(.*?):\/\/.*$/, "$1"),
        3,
        reconcileError.message
      )}`
    );
  }
  return embeddedSourceErrors;
}
