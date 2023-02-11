import { TIMESTAMPS } from "@rilldata/web-common/lib/duckdb-data-types";
import type {
  StructTypeField,
  V1StructType,
} from "@rilldata/web-common/runtime-client";

// This file has simple code that will eventually be moved into selectors similar to other entities

const isFieldColumnATimestamp = (field: StructTypeField) =>
  TIMESTAMPS.has(field.type.code as string);

export const selectTimestampColumnFromSchema = (schema: V1StructType) =>
  (schema?.fields?.filter(isFieldColumnATimestamp) ?? []).map((f) => f.name);
