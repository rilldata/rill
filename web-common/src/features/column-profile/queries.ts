import {
  createQueryServiceColumnCardinality,
  createQueryServiceColumnNullCount,
  createQueryServiceColumnNumericHistogram,
  createQueryServiceColumnRollupInterval,
  createQueryServiceColumnTimeGrain,
  createQueryServiceColumnTimeSeries,
  createQueryServiceColumnTopK,
  createQueryServiceTableCardinality,
  QueryServiceColumnNumericHistogramHistogramMethod,
  type V1ProfileColumn,
  type V1TableColumnsResponse,
  type V1TimeSeriesValue,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import {
  getPriority,
  getPriorityForColumn,
} from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
import {
  keepPreviousData,
  type QueryObserverResult,
} from "@tanstack/query-core";
import { derived, type Readable, writable } from "svelte/store";

export function isFetching(...queries) {
  return queries.some((query) => query?.isFetching);
}

export type ColumnSummary = V1ProfileColumn & {
  nullCount?: number;
  cardinality?: number;
  isFetching: boolean;
};

/** for each entry in a profile column results, return the null count and the column cardinality */
export function getSummaries(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  objectName: string,
  profileColumnResponse: QueryObserverResult<V1TableColumnsResponse, HTTPError>,
): Readable<Array<ColumnSummary>> {
  return derived(
    profileColumnResponse?.data?.profileColumns?.map((column) => {
      return derived(
        [
          writable(column),
          createQueryServiceColumnNullCount(
            instanceId,
            objectName,
            {
              connector,
              database,
              databaseSchema,
              columnName: column.name,
            },
            {
              query: {
                placeholderData: keepPreviousData,
                enabled: !profileColumnResponse.isFetching,
              },
            },
          ),
          createQueryServiceColumnCardinality(
            instanceId,
            objectName,
            {
              connector,
              database,
              databaseSchema,
              columnName: column.name,
            },
            {
              query: {
                placeholderData: keepPreviousData,
                enabled: !profileColumnResponse.isFetching,
              },
            },
          ),
        ],
        ([col, nullValues, cardinality]) => {
          return {
            ...col,
            nullCount: nullValues?.data?.count,
            cardinality: cardinality?.data?.categoricalSummary?.cardinality,
            isFetching:
              profileColumnResponse.isFetching ||
              nullValues?.isFetching ||
              cardinality?.isFetching,
          };
        },
      );
    }) ?? [],

    (combos) => {
      return combos;
    },
  );
}

export function getNullPercentage(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  objectName: string,
  columnName: string,
  enabled = true,
) {
  const nullQuery = createQueryServiceColumnNullCount(
    instanceId,
    objectName,
    {
      connector,
      database,
      databaseSchema,
      columnName,
    },
    {
      query: {
        enabled,
      },
    },
  );
  const totalRowsQuery = createQueryServiceTableCardinality(
    instanceId,
    objectName,
    {
      connector,
      database,
      databaseSchema,
    },
    {
      query: {
        enabled,
      },
    },
  );
  return derived([nullQuery, totalRowsQuery], ([nulls, totalRows]) => {
    return {
      // SAFETY: `.count` should presumably always exist in a
      // null count query response
      nullCount: nulls?.data?.count as number,
      // SAFETY: `.cardinality` should presumably always exist
      // in a carnality query response, but it's not typed as required.
      totalRows: +(totalRows?.data?.cardinality as string),
      isFetching: nulls?.isFetching || totalRows?.isFetching,
    };
  });
}

export function getCountDistinct(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  objectName: string,
  columnName: string,
  enabled = true,
) {
  const cardinalityQuery = createQueryServiceColumnCardinality(
    instanceId,
    objectName,
    { connector, database, databaseSchema, columnName },
    {
      query: {
        enabled,
      },
    },
  );

  const totalRowsQuery = createQueryServiceTableCardinality(
    instanceId,
    objectName,
    { connector, database, databaseSchema },
    {
      query: {
        enabled,
      },
    },
  );

  return derived(
    [cardinalityQuery, totalRowsQuery],
    ([cardinality, totalRows]) => {
      return {
        cardinality: cardinality?.data?.categoricalSummary?.cardinality,
        // SAFETY: if the V1TableCardinalityResponse exists in `totalRows`,
        // then `.cardinality` should presumably always exist in the
        // cardinality query response, so we should be able to cast it to
        // a string. It's just not typed as required b/c of protobuf limitations.
        totalRows: +(totalRows?.data?.cardinality as string),
        isFetching: cardinality?.isFetching || totalRows?.isFetching,
      };
    },
  );
}

export function getTopK(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  objectName: string,
  columnName: string,
  enabled = true,
  active = false,
) {
  const topKQuery = createQueryServiceColumnTopK(
    instanceId,
    objectName,
    {
      connector,
      database,
      databaseSchema,
      columnName: columnName,
      agg: "count(*)",
      k: 75,
      priority: getPriorityForColumn("topk", active),
    },
    {
      query: {
        enabled,
      },
    },
  );
  return derived(topKQuery, ($topKQuery) => {
    return $topKQuery?.data?.categoricalSummary?.topK?.entries;
  });
}

function convertPoint(point: V1TimeSeriesValue) {
  const next = {
    ...point,
    count: point?.records?.count as number,
    ts: point.ts ? new Date(point.ts) : new Date(0),
  };
  if (next.count == null || !isFinite(next.count)) {
    next.count = 0;
  }

  return next;
}

export type TimestampDataPoint = ReturnType<typeof convertPoint>;

export function getTimeSeriesAndSpark(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  objectName: string,
  columnName: string,
  enabled = true,
  active = false,
) {
  const query = createQueryServiceColumnTimeSeries(
    instanceId,
    objectName,
    // FIXME: convert pixel back to number once the API
    {
      connector,
      database,
      databaseSchema,
      timestampColumnName: columnName,
      measures: [
        {
          expression: "count(*)",
        },
      ],
      pixels: 92,
      priority: getPriorityForColumn("timeseries", active),
    },
    {
      query: { enabled },
    },
  );
  const estimatedInterval = createQueryServiceColumnRollupInterval(
    instanceId,
    objectName,
    {
      connector,
      database,
      databaseSchema,
      columnName,
      priority: getPriorityForColumn("rollup-interval", active),
    },
    {
      query: {
        enabled,
      },
    },
  );

  const smallestTimeGrain = createQueryServiceColumnTimeGrain(
    instanceId,
    objectName,
    {
      connector,
      database,
      databaseSchema,
      columnName,
      priority: getPriorityForColumn("smallest-time-grain", active),
    },
    {
      query: {
        enabled,
      },
    },
  );

  return derived(
    [query, estimatedInterval, smallestTimeGrain],
    ([$query, $estimatedInterval, $smallestTimeGrain]) => {
      const data = $query?.data?.rollup?.results?.map(convertPoint) || [];

      const spark = $query?.data?.rollup?.spark?.map(convertPoint) || [];
      return {
        isFetching: $query?.isFetching,
        estimatedRollupInterval: $estimatedInterval?.data,
        smallestTimegrain: $smallestTimeGrain?.data?.timeGrain,
        data,
        spark,
      };
    },
  );
}

export function getNumericHistogram(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  objectName: string,
  columnName: string,
  histogramMethod: QueryServiceColumnNumericHistogramHistogramMethod,
  enabled = true,
) {
  return createQueryServiceColumnNumericHistogram(
    instanceId,
    objectName,
    {
      connector,
      database,
      databaseSchema,
      columnName,
      histogramMethod,
      priority: getPriority("numeric-histogram"),
    },
    {
      query: {
        select(query) {
          return query?.numericSummary?.numericHistogramBins?.bins;
        },
        enabled,
      },
    },
  );
}
