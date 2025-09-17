import type {
  Schema as MetricsResolverQuery,
  TimeRange,
  Sort,
} from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

/**
 * Types and validation for metrics view queries used in the open-query functionality.
 * This defines the structure of queries that can be passed via URL parameters to open dashboards.
 */

/**
 * Validates and normalizes a raw query object into a proper Query type.
 * Throws an error if the query is invalid.
 */
export function validateQuery(
  rawQuery: MetricsResolverQuery,
): MetricsResolverQuery {
  if (!rawQuery || typeof rawQuery !== "object") {
    throw new Error("Query must be an object");
  }

  if (!rawQuery.metrics_view || typeof rawQuery.metrics_view !== "string") {
    throw new Error("metrics_view is required and must be a string");
  }

  const query: MetricsResolverQuery = {
    metrics_view: rawQuery.metrics_view,
  };

  // Validate dimensions
  if (rawQuery.dimensions !== undefined) {
    if (!Array.isArray(rawQuery.dimensions)) {
      throw new Error("dimensions must be an array");
    }
    query.dimensions = rawQuery.dimensions.map((dim, index) => {
      if (!dim || typeof dim !== "object") {
        throw new Error(`dimensions[${index}] must be an object`);
      }
      if (!dim.name || typeof dim.name !== "string") {
        throw new Error(
          `dimensions[${index}].name is required and must be a string`,
        );
      }
      return { name: dim.name };
    });
  }

  // Validate measures
  if (rawQuery.measures !== undefined) {
    if (!Array.isArray(rawQuery.measures)) {
      throw new Error("measures must be an array");
    }
    query.measures = rawQuery.measures.map((measure, index) => {
      if (!measure || typeof measure !== "object") {
        throw new Error(`measures[${index}] must be an object`);
      }
      if (!measure.name || typeof measure.name !== "string") {
        throw new Error(
          `measures[${index}].name is required and must be a string`,
        );
      }
      // TODO: validate compute
      return { name: measure.name, compute: measure.compute };
    });
  }

  // Validate time_range
  if (rawQuery.time_range !== undefined) {
    query.time_range = validateTimeRange(rawQuery.time_range, "time_range");
  }

  // Validate comparison_time_range
  if (rawQuery.comparison_time_range !== undefined) {
    query.comparison_time_range = validateTimeRange(
      rawQuery.comparison_time_range,
      "comparison_time_range",
    );
  }

  // Validate where (allow any structure for now as it's an Expression type)
  if (rawQuery.where !== undefined) {
    query.where = rawQuery.where;
  }

  // Validate sort
  if (rawQuery.sort !== undefined) {
    if (!Array.isArray(rawQuery.sort)) {
      throw new Error("sort must be an array");
    }
    query.sort = rawQuery.sort.map((sortItem, index) => {
      if (!sortItem || typeof sortItem !== "object") {
        throw new Error(`sort[${index}] must be an object`);
      }
      if (!sortItem.name || typeof sortItem.name !== "string") {
        throw new Error(`sort[${index}].name is required and must be a string`);
      }
      const sort: Sort = { name: sortItem.name };
      if (sortItem.desc !== undefined) {
        if (typeof sortItem.desc !== "boolean") {
          throw new Error(`sort[${index}].desc must be a boolean`);
        }
        sort.desc = sortItem.desc;
      }
      return sort;
    });
  }

  // Validate time_zone
  if (rawQuery.time_zone !== undefined) {
    if (typeof rawQuery.time_zone !== "string") {
      throw new Error("time_zone must be a string");
    }
    query.time_zone = rawQuery.time_zone;
  }

  return query;
}

const TimeRangeKeysToValidate: (keyof TimeRange)[] = [
  "start",
  "end",
  "expression",
  "iso_duration",
  "iso_offset",
];
function validateTimeRange(timeRange: TimeRange, property: string): TimeRange {
  if (typeof timeRange !== "object") {
    throw new Error(`${property} must be an object`);
  }
  const validTimeRange: TimeRange = {};

  TimeRangeKeysToValidate.forEach((key) => {
    if (timeRange[key] === undefined) return;
    if (typeof timeRange[key] !== "string") {
      throw new Error(`${property}.${key} must be a string`);
    }
    validTimeRange[key] = timeRange[key];
  });

  return validTimeRange;
}
