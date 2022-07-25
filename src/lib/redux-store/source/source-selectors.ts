import type { PersistentTableEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type { DerivedTableEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { TIMESTAMPS } from "$lib/duckdb-data-types";

// Source doesn't have a slice as of now.
// This file has simple code that will eventually be moved into selectors similar to other entities

export const derivedSourceHasTimestampColumn = (
  derivedSource: DerivedTableEntity
) => derivedSource.profile.some((column) => TIMESTAMPS.has(column.type));

export const selectSourcesWithTimestampColumns = (
  persistentSources: Array<PersistentTableEntity>,
  derivedSources: Array<DerivedTableEntity>
): Array<PersistentTableEntity> => {
  return derivedSources
    .filter(derivedSourceHasTimestampColumn)
    .map((derivedSource) =>
      persistentSources.find(
        (persistentSource) => (persistentSource.id = derivedSource.id)
      )
    );
};
