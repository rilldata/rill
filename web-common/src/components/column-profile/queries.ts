import { convertTimestampPreview } from "@rilldata/web-common/lib/convertTimestampPreview";
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
  V1ProfileColumn,
  V1TableColumnsResponse,
} from "@rilldata/web-common/runtime-client";
import { getPriorityForColumn } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
import type { QueryObserverResult } from "@tanstack/query-core";
import { derived, Readable, writable } from "svelte/store";

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
  objectName: string,
  instanceId: string,
  profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>,
): Readable<Array<ColumnSummary>> {
  return derived(
    profileColumnResponse?.data?.profileColumns?.map((column) => {
      return derived(
        [
          writable(column),
          createQueryServiceColumnNullCount(
            instanceId,
            objectName,
            { columnName: column.name },
            {
              query: {
                keepPreviousData: true,
                enabled: !profileColumnResponse.isFetching,
              },
            },
          ),
          createQueryServiceColumnCardinality(
            instanceId,
            objectName,
            { columnName: column.name },
            {
              query: {
                keepPreviousData: true,
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
  objectName: string,
  columnName: string,
  enabled = true,
) {
  const nullQuery = createQueryServiceColumnNullCount(
    instanceId,
    objectName,
    {
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
    {},
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
  objectName: string,
  columnName: string,
  enabled = true,
) {
  const cardinalityQuery = createQueryServiceColumnCardinality(
    instanceId,
    objectName,
    { columnName },
    {
      query: { enabled },
    },
  );

  const totalRowsQuery = createQueryServiceTableCardinality(
    instanceId,
    objectName,
    {},
    {
      query: { enabled },
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
  objectName: string,
  columnName: string,
  enabled = true,
  active = false,
) {
  const topKQuery = createQueryServiceColumnTopK(
    instanceId,
    objectName,
    {
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

export function getTimeSeriesAndSpark(
  instanceId: string,
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
    { columnName, priority: getPriorityForColumn("rollup-interval", active) },
    {
      query: { enabled },
    },
  );

  const smallestTimeGrain = createQueryServiceColumnTimeGrain(
    instanceId,
    objectName,
    {
      columnName,
      priority: getPriorityForColumn("smallest-time-grain", active),
    },
    {
      query: { enabled },
    },
  );

  return derived(
    [query, estimatedInterval, smallestTimeGrain],
    ([$query, $estimatedInterval, $smallestTimeGrain]) => {
      return {
        isFetching: $query?.isFetching,
        estimatedRollupInterval: $estimatedInterval?.data,
        smallestTimegrain: $smallestTimeGrain?.data?.timeGrain,
        data: convertTimestampPreview(
          $query?.data?.rollup?.results?.map((di) => {
            const next = { ...di, count: di?.records?.count };
            if (next.count == null || !isFinite(next.count)) {
              next.count = 0;
            }
            return next;
          }) || [],
        ),
        spark: convertTimestampPreview(
          $query?.data?.rollup?.spark?.map((di) => {
            const next = { ...di, count: di?.records?.count };
            if (next.count == null || !isFinite(next.count)) {
              next.count = 0;
            }
            return next;
          }) || [],
        ),
      };
    },
  );
}

export function getNumericHistogram(
  instanceId: string,
  objectName: string,
  columnName: string,
  histogramMethod: QueryServiceColumnNumericHistogramHistogramMethod,
  enabled = true,
  active = false,
) {
  return createQueryServiceColumnNumericHistogram(
    instanceId,
    objectName,
    {
      columnName,
      histogramMethod,
      priority: getPriorityForColumn("numeric-histogram", active),
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
