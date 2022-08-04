import type { DataProfileEntity } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import { TIMESTAMPS } from "$lib/duckdb-data-types";
import type { ProfileColumn } from "$lib/types";

// Source doesn't have a slice as of now.
// This file has simple code that will eventually be moved into selectors similar to other entities

const isProfileColumnATimestamp = (column: ProfileColumn) =>
  TIMESTAMPS.has(column.type);

export const derivedProfileEntityHasTimestampColumn = (
  derivedProfileEntity: DataProfileEntity
) => derivedProfileEntity.profile.some(isProfileColumnATimestamp);

export const selectTimestampColumnFromProfileEntity = (
  derivedProfileEntity: DataProfileEntity
) => derivedProfileEntity?.profile?.filter(isProfileColumnATimestamp) ?? [];
