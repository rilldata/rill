import { TIMESTAMPS } from "@rilldata/web-common/lib/duckdb-data-types";
import type {
  StructTypeField,
  V1StructType,
} from "@rilldata/web-common/runtime-client";

// Source doesn't have a slice as of now.
// This file has simple code that will eventually be moved into selectors similar to other entities

const isFieldColumnATimestamp = (field: StructTypeField) => {
  return TIMESTAMPS.has(field.type.code as string);
};

export const schemaHasTimestampColumn = (schema: V1StructType) => {
  return schema?.fields?.some(isFieldColumnATimestamp);
};

export const selectTimestampColumnFromSchema = (schema: V1StructType) =>
  (schema?.fields?.filter(isFieldColumnATimestamp) ?? []).map((f) => f.name);
