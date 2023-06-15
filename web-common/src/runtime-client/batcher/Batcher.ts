import { V1QueryBatchType } from "@rilldata/web-common/runtime-client";
import { batchRequest } from "@rilldata/web-common/runtime-client/batcher/batchRequest";
import type { QueryEntry } from "@rilldata/web-common/runtime-client/batcher/batchRequest";
import { fetchWrapper } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { FetchWrapperOptions } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { getPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";

// Examples:
// v1/instances/id/queries/columns-profile/tables/table-name
// v1/instances/id/queries/metrics-views/mv-name/timeseries
export const BatcherUrlExtractorRegex =
  /v1\/instances\/[\w-]+\/queries\/([\w-]+)\/([\w-]+)\/([\w-]+)/;

const ProfileQueryMap: Record<string, V1QueryBatchType> = {
  // TODO: metrics view
  "rollup-interval": V1QueryBatchType.ColumnRollupInterval,
  topk: V1QueryBatchType.ColumnTopK,
  "null-count": V1QueryBatchType.ColumnNullCount,
  "descriptive-statistics": V1QueryBatchType.ColumnDescriptiveStatistics,
  "smallest-time-grain": V1QueryBatchType.ColumnTimeGrain,
  "numeric-histogram": V1QueryBatchType.ColumnNumericHistogram,
  "rug-histogram": V1QueryBatchType.ColumnRugHistogram,
  "time-range-summary": V1QueryBatchType.ColumnTimeRange,
  "column-cardinality": V1QueryBatchType.ColumnCardinality,
  timeseries: V1QueryBatchType.ColumnTimeSeries,
  "table-cardinality": V1QueryBatchType.TableCardinality,
  "columns-profile": V1QueryBatchType.TableColumns,
  rows: V1QueryBatchType.TableRows,
};

export class Batcher {
  private queries = new Array<QueryEntry>();
  private instanceId: string;
  private timer: number;
  private baseUrl: string;

  public setInstanceId(instanceId: string) {
    this.instanceId = instanceId;
  }

  public add(requestOptions: FetchWrapperOptions) {
    // prepend after parsing to make parsing faster
    requestOptions.url = `${requestOptions.baseUrl}${requestOptions.url}`;
    // TODO: support multiple runtimes for cloud
    this.baseUrl = requestOptions.baseUrl;

    const urlMatch = BatcherUrlExtractorRegex.exec(requestOptions.url);
    if (!urlMatch) return fetchWrapper(requestOptions);

    const profileType = urlMatch[1];
    if (!ProfileQueryMap[profileType]) return fetchWrapper(requestOptions);

    const name = urlMatch[3];

    let priority: number;
    const request = requestOptions.params ?? requestOptions.data ?? {};
    requestOptions.params ??= {};
    priority =
      requestOptions.data?.priority ??
      (requestOptions.params.priority as number);

    if (!priority) {
      priority = getPriority(profileType);
    }
    request.priority = priority;
    request.instanceId = this.instanceId;
    request.tableName = name;

    return new Promise((resolve, reject) => {
      this.queries.push([
        ProfileQueryMap[profileType],
        request,
        resolve,
        reject,
        requestOptions.signal,
      ]);
      this.throttleBatch();
    });
  }

  private throttleBatch() {
    if (!this.timer) {
      this.timer = setTimeout(() => {
        this.timer = 0;
        batchRequest(
          `${this.baseUrl}/v1/instances/${this.instanceId}/query/batch`,
          this.queries
        );
        this.queries = [];
      }, 100) as any;
    }
  }
}
