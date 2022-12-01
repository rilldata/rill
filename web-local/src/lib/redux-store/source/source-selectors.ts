import type {
  StructTypeField,
  V1StructType,
} from "@rilldata/web-common/runtime-client";
import type { DataProfileEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import { TIMESTAMPS } from "../../duckdb-data-types";
import type { ProfileColumn } from "../../types";

// Source doesn't have a slice as of now.
// This file has simple code that will eventually be moved into selectors similar to other entities

const isProfileColumnATimestamp = (column: ProfileColumn) =>
  TIMESTAMPS.has(column.type);

const isFieldColumnATimestamp = (field: StructTypeField) =>
  TIMESTAMPS.has(field.type.code as string);

export const derivedProfileEntityHasTimestampColumn = (
  derivedProfileEntity: DataProfileEntity
) => derivedProfileEntity?.profile.some(isProfileColumnATimestamp);

export const selectTimestampColumnFromProfileEntity = (
  derivedProfileEntity: DataProfileEntity
) => derivedProfileEntity?.profile?.filter(isProfileColumnATimestamp) ?? [];

export const schemaHasTimestampColumn = (schema: V1StructType) =>
  schema?.fields?.some(isFieldColumnATimestamp);

export const selectTimestampColumnFromModelSchema = (schema: V1StructType) =>
  (schema?.fields?.filter(isFieldColumnATimestamp) ?? []).map((f) => f.name);
