import {
  RuntimeServiceGetNumericHistogramHistogramMethod,
  useRuntimeServiceEstimateRollupInterval,
  useRuntimeServiceEstimateSmallestTimeGrain,
  useRuntimeServiceGenerateTimeSeries,
  useRuntimeServiceGetCardinalityOfColumn,
  useRuntimeServiceGetNullCount,
  useRuntimeServiceGetNumericHistogram,
  useRuntimeServiceGetTableCardinality,
  useRuntimeServiceGetTopK,
  V1ProfileColumn,
} from "@rilldata/web-common/runtime-client";
import { getPriorityForColumn } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
import { convertTimestampPreview } from "@rilldata/web-local/lib/util/convertTimestampPreview";
import { derived, Readable, writable } from "svelte/store";

export function isFetching(...queries) {
  return queries.some((query) => query?.isFetching);
}

export type ColumnSummary = V1ProfileColumn & {
  nullCount: number;
  cardinality: number;
  isFetching: boolean;
};

/** for each entry in a profile column results, return the null count and the column cardinality */
export function getSummaries(
  objectName: string,
  instanceId: string,
  profileColumnResults: Array<V1ProfileColumn>
): Readable<Array<ColumnSummary>> {
  if (!profileColumnResults && !profileColumnResults?.length) return;
  return derived(
    profileColumnResults.map((column) => {
      return derived(
        [
          writable(column),
          useRuntimeServiceGetNullCount(
            instanceId,
            objectName,
            { columnName: column.name },
            {
              query: { keepPreviousData: true },
            }
          ),
          useRuntimeServiceGetCardinalityOfColumn(
            instanceId,
            objectName,
            { columnName: column.name },
            { query: { keepPreviousData: true } }
          ),
        ],
        ([col, nullValues, cardinality]) => {
          return {
            ...col,
            nullCount: +nullValues?.data?.count,
            cardinality: +cardinality?.data?.categoricalSummary?.cardinality,
            isFetching: nullValues?.isFetching || cardinality?.isFetching,
          };
        }
      );
    }),

    (combos) => {
      return combos;
    }
  );
}

export function getNullPercentage(
  instanceId: string,
  objectName: string,
  columnName: string
) {
  const nullQuery = useRuntimeServiceGetNullCount(instanceId, objectName, {
    columnName,
  });
  const totalRowsQuery = useRuntimeServiceGetTableCardinality(
    instanceId,
    objectName
  );
  return derived([nullQuery, totalRowsQuery], ([nulls, totalRows]) => {
    return {
      nullCount: nulls?.data?.count,
      totalRows: +totalRows?.data?.cardinality,
      isFetching: nulls?.isFetching || totalRows?.isFetching,
    };
  });
}

export function getCountDistinct(
  instanceId: string,
  objectName: string,
  columnName: string
) {
  const cardinalityQuery = useRuntimeServiceGetCardinalityOfColumn(
    instanceId,
    objectName,
    { columnName }
  );

  const totalRowsQuery = useRuntimeServiceGetTableCardinality(
    instanceId,
    objectName
  );

  return derived(
    [cardinalityQuery, totalRowsQuery],
    ([cardinality, totalRows]) => {
      return {
        cardinality: cardinality?.data?.categoricalSummary?.cardinality,
        totalRows: +totalRows?.data?.cardinality,
        isFetching: cardinality?.isFetching || totalRows?.isFetching,
      };
    }
  );
}

export function getTopK(
  instanceId: string,
  objectName: string,
  columnName: string,
  active = false
) {
  const topKQuery = useRuntimeServiceGetTopK(instanceId, objectName, {
    columnName: columnName,
    agg: "count(*)",
    k: 75,
    priority: getPriorityForColumn("topk", active),
  });
  return derived(topKQuery, ($topKQuery) => {
    return $topKQuery?.data?.categoricalSummary?.topK?.entries;
  });
}

export function getTimeSeriesAndSpark(
  instanceId: string,
  objectName: string,
  columnName: string,
  active = false
) {
  const query = useRuntimeServiceGenerateTimeSeries(
    instanceId,
    objectName,
    // FIXME: convert pixel back to number once the API
    {
      timestampColumnName: columnName,
      pixels: 92,
      priority: getPriorityForColumn("timeseries", active),
    }
  );
  const estimatedInterval = useRuntimeServiceEstimateRollupInterval(
    instanceId,
    objectName,
    { columnName, priority: getPriorityForColumn("rollup-interval", active) }
  );

  const smallestTimeGrain = useRuntimeServiceEstimateSmallestTimeGrain(
    instanceId,
    objectName,
    {
      columnName,
      priority: getPriorityForColumn("smallest-time-grain", active),
    }
  );

  return derived(
    [query, estimatedInterval, smallestTimeGrain],
    ([$query, $estimatedInterval, $smallestTimeGrain]) => {
      return {
        isFetching: $query?.isFetching,
        estimatedRollupInterval: $estimatedInterval?.data,
        smallestTimegrain: $smallestTimeGrain?.data?.timeGrain,
        data: convertTimestampPreview(
          $query?.data?.rollup?.results.map((di) => {
            const next = { ...di, count: di.records.count };
            if (next.count == null || !isFinite(next.count)) {
              next.count = 0;
            }
            return next;
          }) || []
        ),
        spark: convertTimestampPreview(
          $query?.data?.rollup?.spark.map((di) => {
            const next = { ...di, count: di.records.count };
            if (next.count == null || !isFinite(next.count)) {
              next.count = 0;
            }
            return next;
          }) || []
        ),
      };
    }
  );
}

export function getNumericHistogram(
  instanceId: string,
  objectName: string,
  columnName: string,
  histogramMethod: RuntimeServiceGetNumericHistogramHistogramMethod,
  active = false
) {
  return useRuntimeServiceGetNumericHistogram(
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
      },
    }
  );
}
