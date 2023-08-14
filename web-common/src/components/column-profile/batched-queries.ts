import {
  createQueryServiceColumnCardinality,
  createQueryServiceColumnDescriptiveStatistics,
  createQueryServiceColumnNullCount,
  createQueryServiceColumnNumericHistogram,
  createQueryServiceColumnRollupInterval,
  createQueryServiceColumnRugHistogram,
  createQueryServiceColumnTimeGrain,
  createQueryServiceColumnTimeSeries,
  createQueryServiceColumnTopK,
  createQueryServiceTableCardinality,
  QueryServiceColumnNumericHistogramHistogramMethod,
  V1QueryBatchEntry,
  V1QueryBatchResponse,
} from "@rilldata/web-common/runtime-client";
import type { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { getPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
import type { QueryFunction } from "@tanstack/svelte-query";

// TODO: integrate active priority

export function createBatchedColumnNullCount(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnNullCount(
    instanceId,
    tableName,
    { columnName },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnNullCountRequest: {
              instanceId,
              tableName,
              columnName,
            },
          },
          "null-count",
          "columnNullCountResponse"
        ),
      },
    }
  );
}

export function createBatchedColumnCardinality(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnCardinality(
    instanceId,
    tableName,
    { columnName },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnCardinalityRequest: {
              instanceId,
              tableName,
              columnName,
            },
          },
          "column-cardinality",
          "columnCardinalityResponse"
        ),
      },
    }
  );
}

export function createBatchedColumnTopKQuery(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnTopK(
    instanceId,
    tableName,
    {
      columnName,
      // TODO: keep these in a common place
      agg: "count(*)",
      k: 75,
    },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnTopKRequest: {
              instanceId,
              tableName,
              columnName,
              agg: "count(*)",
              k: 75,
            },
          },
          "topk",
          "columnTopKResponse"
        ),
      },
    }
  );
}

export function createBatchedColumnTimeSeriesQuery(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnTimeSeries(
    instanceId,
    tableName,
    {
      timestampColumnName: columnName,
      // TODO: keep these in a common place
      measures: [{ expression: "count(*)" }],
      pixels: 92,
    },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnTimeSeriesRequest: {
              instanceId,
              tableName,
              timestampColumnName: columnName,
              measures: [{ expression: "count(*)" }],
              pixels: 92,
            },
          },
          "timeseries",
          "columnTimeSeriesResponse"
        ),
      },
    }
  );
}

export function createBatchedServiceColumnRollupIntervalQuery(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnRollupInterval(
    instanceId,
    tableName,
    { columnName },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnRollupIntervalRequest: {
              instanceId,
              tableName,
              columnName,
            },
          },
          "rollup-interval",
          "columnRollupIntervalResponse"
        ),
      },
    }
  );
}

export function createBatchedColumnTimeGrainQuery(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnTimeGrain(
    instanceId,
    tableName,
    { columnName },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnTimeGrainRequest: {
              instanceId,
              tableName,
              columnName,
            },
          },
          "smallest-time-grain",
          "columnTimeGrainResponse"
        ),
      },
    }
  );
}

export function createBatchedColumnNumericHistogramQuery(
  instanceId: string,
  tableName: string,
  columnName: string,
  histogramMethod: QueryServiceColumnNumericHistogramHistogramMethod,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnNumericHistogram(
    instanceId,
    tableName,
    {
      columnName,
      histogramMethod,
    },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnNumericHistogramRequest: {
              instanceId,
              tableName,
              columnName,
              histogramMethod,
            },
          },
          "numeric-histogram",
          "columnNumericHistogramResponse"
        ),
      },
    }
  );
}

export function createBatchedColumnRugHistogramQuery(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnRugHistogram(
    instanceId,
    tableName,
    { columnName },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnRugHistogramRequest: {
              instanceId,
              tableName,
              columnName,
            },
          },
          "rug-histogram",
          "columnRugHistogramResponse"
        ),
      },
    }
  );
}

export function createBatchedColumnDescriptiveStatisticsQuery(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceColumnDescriptiveStatistics(
    instanceId,
    tableName,
    {
      columnName: columnName,
    },
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            columnDescriptiveStatisticsRequest: {
              instanceId,
              tableName,
              columnName,
            },
          },
          "descriptive-statistics",
          "columnDescriptiveStatisticsResponse"
        ),
      },
    }
  );
}

export function createBatchedTableCardinalityQuery(
  instanceId: string,
  tableName: string,
  batchedRequest: BatchedRequest
) {
  return createQueryServiceTableCardinality(
    instanceId,
    tableName,
    {},
    {
      query: {
        queryFn: wrapQueryFunction(
          batchedRequest,
          {
            tableCardinalityRequest: {
              instanceId,
              tableName,
            },
          },
          "table-cardinality",
          "tableCardinalityResponse"
        ),
      },
    }
  );
}

function wrapQueryFunction(
  batchedRequest: BatchedRequest,
  request: V1QueryBatchEntry,
  type: string,
  responseKey: keyof V1QueryBatchResponse
): QueryFunction {
  batchedRequest.register();
  return ({ signal }) => {
    return new Promise((resolve, reject) => {
      const priority = getPriority(type);
      batchedRequest[Object.keys(batchedRequest)[0]].priority = priority;
      batchedRequest.add(
        request,
        priority,
        (data) => {
          resolve(data?.[responseKey]);
        },
        reject,
        signal
      );
    });
  };
}
