import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/fetch-streaming-wrapper";
import type {
  QueryServiceQueryBatchBody,
  V1Query,
  V1QueryBatchResponse,
  V1QueryResult,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

type BatchQueryEntry<Type extends BatchQueryTypes> = {
  query: V1Query;
  signal: AbortSignal | undefined;
  resolve: (resp: V1QueryResult[`${Type}Response`], index: number) => void;
  reject: (err: Error) => void;
  type: BatchQueryTypes;
};

/**
 * Extracts the query types supported in BatchQuery.
 */
type BatchQueryTypes = keyof V1Query extends `${infer Args}Request`
  ? Args
  : never;

export class StreamingQueryBatch {
  private timer: ReturnType<typeof setTimeout> | undefined;
  private queryEntries: BatchQueryEntry<any>[] = [];

  public constructor(private readonly batchPeriod: number) {}

  public fetch<Type extends BatchQueryTypes>(
    type: Type,
    request: V1Query[`${Type}Request`],
    signal: AbortSignal | undefined,
  ): Promise<NonNullable<V1QueryResult[`${Type}Response`]>> {
    return new Promise((resolve, reject) => {
      this.queryEntries.push({
        query: {
          [`${type}Request`]: request,
        },
        signal,
        resolve,
        reject,
        type,
      });

      if (this.timer) return;
      this.timer = setTimeout(() => {
        void this.runBatch();
        this.timer = undefined;
      }, this.batchPeriod);
    });
  }

  private async runBatch() {
    const runtimeState = get(runtime);
    const queries = this.queryEntries;
    this.queryEntries = [];

    const headers = { "Content-Type": "application/json" };
    const jwt = runtimeState.jwt;
    if (jwt) {
      headers["Authorization"] = `Bearer ${jwt.token}`;
    }

    const body: QueryServiceQueryBatchBody = {
      queries: [],
    };
    // create a single abort controller that cancels if any signal is cancelled.
    const controller = new AbortController();
    queries.forEach(({ query, signal }) => {
      body.queries?.push(query);
      signal?.addEventListener("abort", () => {
        controller.abort(signal.reason || "Query cancelled");
      });
    });

    const stream = streamingFetchWrapper<{ result: V1QueryBatchResponse }>(
      `${runtimeState.host}/v1/instances/${runtimeState.instanceId}/query/batch`,
      "POST",
      body,
      headers,
      controller.signal,
    );

    const seen = new Set<number>();
    for await (const { result } of stream) {
      if (result?.index === undefined || !queries[result.index]) continue;
      // index matches the request order. check batch_query.go
      seen.add(result.index);
      const entry = queries[result.index];
      if (result.error) {
        entry.reject(new Error(result.error));
      } else if (result.result) {
        entry.resolve(result.result[`${entry.type}Response`], result.index);
      } else {
        entry.reject(new Error("unknown error"));
      }
    }

    // after the request if finished reject any missing responses
    queries.forEach(({ reject }, i) => {
      if (seen.has(i)) return;
      reject(new Error("no response"));
    });
  }
}
