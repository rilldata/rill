import type {
  V1ColumnCardinalityRequest,
  V1ColumnDescriptiveStatisticsRequest,
  V1ColumnNullCountRequest,
  V1ColumnNumericHistogramRequest,
  V1ColumnRollupIntervalRequest,
  V1ColumnRugHistogramRequest,
  V1ColumnTimeGrainRequest,
  V1ColumnTimeRangeRequest,
  V1ColumnTimeSeriesRequest,
  V1ColumnTopKRequest,
  V1MetricsViewTimeSeriesRequest,
  V1MetricsViewToplistRequest,
  V1MetricsViewTotalsRequest,
  V1QueryBatchResponse,
  V1QueryBatchSingleRequest,
  V1QueryBatchType,
  V1TableCardinalityRequest,
  V1TableColumnsRequest,
  V1TableRowsRequest,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/streamingFetchWrapper";

export type QueryRequestTypes =
  | V1MetricsViewToplistRequest
  | V1MetricsViewTimeSeriesRequest
  | V1MetricsViewTotalsRequest
  | V1ColumnRollupIntervalRequest
  | V1ColumnTopKRequest
  | V1ColumnNullCountRequest
  | V1ColumnDescriptiveStatisticsRequest
  | V1ColumnTimeGrainRequest
  | V1ColumnNumericHistogramRequest
  | V1ColumnRugHistogramRequest
  | V1ColumnTimeRangeRequest
  | V1ColumnCardinalityRequest
  | V1ColumnTimeSeriesRequest
  | V1TableCardinalityRequest
  | V1TableColumnsRequest
  | V1TableRowsRequest;

export type QueryEntry = [
  type: V1QueryBatchType,
  request: QueryRequestTypes,
  resolve: (data: any) => void,
  reject: (err: Error) => void,
  signal: AbortSignal | undefined
];

export async function batchRequest(url: string, queries: Array<QueryEntry>) {
  const request = {
    queries: queries.map(([type, req], index) => mapRequest(index, type, req)),
  };
  const controller = new AbortController();
  const stream = streamingFetchWrapper<{ result: V1QueryBatchResponse }>(
    url,
    "post",
    request,
    controller.signal
  );

  queries.forEach(([, , , , signal]) => {
    signal?.addEventListener(
      "abort",
      () => {
        if (controller.signal.aborted) return;
        controller.abort();
        stream.throw(new Error("cancelled"));
      },
      {
        once: true,
      }
    );
  });

  const hit = new Set<number>();

  for await (const res of stream) {
    const idx = res.result.id ?? 0;
    hit.add(idx);
    if (res.result.error) {
      queries[idx][3](new Error(res.result.error));
      continue;
    }
    queries[idx][2](mapResponse(res.result));
  }

  for (let i = 0; i < queries.length; i++) {
    if (hit.has(i)) continue;
    queries[i][3](new Error("No response"));
  }
}

function mapRequest(
  id: number,
  type: V1QueryBatchType,
  req: QueryRequestTypes
) {
  const batchReq: V1QueryBatchSingleRequest = {
    id,
    type,
  };
  switch (type) {
    case "MetricsViewToplist":
      batchReq.metricsViewToplistRequest = req as V1MetricsViewToplistRequest;
      break;
    case "MetricsViewTimeSeries":
      batchReq.metricsViewTimeSeriesRequest = req;
      break;
    case "MetricsViewTotals":
      batchReq.metricsViewTotalsRequest = req;
      break;
    case "ColumnRollupInterval":
      batchReq.columnRollupIntervalRequest = req;
      break;
    case "ColumnTopK":
      batchReq.columnTopKRequest = req;
      break;
    case "ColumnNullCount":
      batchReq.columnNullCountRequest = req;
      break;
    case "ColumnDescriptiveStatistics":
      batchReq.columnDescriptiveStatisticsRequest = req;
      break;
    case "ColumnTimeGrain":
      batchReq.columnTimeGrainRequest = req;
      break;
    case "ColumnNumericHistogram":
      batchReq.columnNumericHistogramRequest = req;
      break;
    case "ColumnRugHistogram":
      batchReq.columnRugHistogramRequest = req;
      break;
    case "ColumnTimeRange":
      batchReq.columnTimeRangeRequest = req;
      break;
    case "ColumnCardinality":
      batchReq.columnCardinalityRequest = req;
      break;
    case "ColumnTimeSeries":
      batchReq.columnTimeSeriesRequest = req;
      break;
    case "TableCardinality":
      batchReq.tableCardinalityRequest = req;
      break;
    case "TableColumns":
      batchReq.tableColumnsRequest = req;
      break;
    case "TableRows":
      batchReq.tableRowsRequest = req as V1TableRowsRequest;
      break;
  }
  return batchReq;
}

function mapResponse(res: V1QueryBatchResponse) {
  return (
    res.metricsViewToplistResponse ??
    res.metricsViewTimeSeriesResponse ??
    res.metricsViewTotalsResponse ??
    res.columnRollupIntervalResponse ??
    res.columnTopKResponse ??
    res.columnNullCountResponse ??
    res.columnDescriptiveStatisticsResponse ??
    res.columnTimeGrainResponse ??
    res.columnNumericHistogramResponse ??
    res.columnRugHistogramResponse ??
    res.columnTimeRangeResponse ??
    res.columnCardinalityResponse ??
    res.columnTimeSeriesResponse ??
    res.tableCardinalityResponse ??
    res.tableColumnsResponse ??
    res.tableRowsResponse
  );
}
