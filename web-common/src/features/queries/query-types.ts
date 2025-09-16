/**
 * Types and validation for metrics view queries used in the open-query functionality.
 * This defines the structure of queries that can be passed via URL parameters to open dashboards.
 */

export interface QueryDimension {
  name: string;
}

export interface QueryMeasure {
  name: string;
}

export interface QueryTimeRange {
  start?: string;
  end?: string;
}

export interface QuerySort {
  name: string;
  desc?: boolean;
}

export interface Query {
  metrics_view: string;
  dimensions?: QueryDimension[];
  measures?: QueryMeasure[];
  time_range?: QueryTimeRange;
  where?: any; // Expression type from proto
  sort?: QuerySort[];
  time_zone?: string;
}

/**
 * Validates and normalizes a raw query object into a proper Query type.
 * Throws an error if the query is invalid.
 */
export function validateQuery(rawQuery: any): Query {
  if (!rawQuery || typeof rawQuery !== "object") {
    throw new Error("Query must be an object");
  }

  if (!rawQuery.metrics_view || typeof rawQuery.metrics_view !== "string") {
    throw new Error("metrics_view is required and must be a string");
  }

  const query: Query = {
    metrics_view: rawQuery.metrics_view,
  };

  // Validate dimensions
  if (rawQuery.dimensions !== undefined) {
    if (!Array.isArray(rawQuery.dimensions)) {
      throw new Error("dimensions must be an array");
    }
    query.dimensions = rawQuery.dimensions.map((dim: any, index: number) => {
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
    query.measures = rawQuery.measures.map((measure: any, index: number) => {
      if (!measure || typeof measure !== "object") {
        throw new Error(`measures[${index}] must be an object`);
      }
      if (!measure.name || typeof measure.name !== "string") {
        throw new Error(
          `measures[${index}].name is required and must be a string`,
        );
      }
      return { name: measure.name };
    });
  }

  // Validate time_range
  if (rawQuery.time_range !== undefined) {
    if (!rawQuery.time_range || typeof rawQuery.time_range !== "object") {
      throw new Error("time_range must be an object");
    }
    const timeRange: QueryTimeRange = {};
    if (rawQuery.time_range.start !== undefined) {
      if (typeof rawQuery.time_range.start !== "string") {
        throw new Error("time_range.start must be a string");
      }
      timeRange.start = rawQuery.time_range.start;
    }
    if (rawQuery.time_range.end !== undefined) {
      if (typeof rawQuery.time_range.end !== "string") {
        throw new Error("time_range.end must be a string");
      }
      timeRange.end = rawQuery.time_range.end;
    }
    query.time_range = timeRange;
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
    query.sort = rawQuery.sort.map((sortItem: any, index: number) => {
      if (!sortItem || typeof sortItem !== "object") {
        throw new Error(`sort[${index}] must be an object`);
      }
      if (!sortItem.name || typeof sortItem.name !== "string") {
        throw new Error(`sort[${index}].name is required and must be a string`);
      }
      const sort: QuerySort = { name: sortItem.name };
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
