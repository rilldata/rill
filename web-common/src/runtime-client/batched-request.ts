import type {
  QueryServiceQueryBatchBody,
  V1QueryBatchEntry,
  V1QueryBatchResponse,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { DefaultQueryPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/streaming-fetch-wrapper";

export type BatchRequest = {
  request: V1QueryBatchEntry;
  priority: number;
  resolve: (data: V1QueryBatchResponse) => void;
  reject: (err: Error) => void;
};
export class BatchedRequest {
  private requests = new Array<BatchRequest>();
  private expectedRequests = 0;
  private controller: AbortController;

  public get ready() {
    return this.requests.length >= this.expectedRequests;
  }

  public register() {
    this.expectedRequests++;
  }

  public add<T>(
    request: V1QueryBatchEntry,
    selector: (data: V1QueryBatchResponse) => T
  ) {
    return new Promise<T>((resolve, reject) => {
      request.key = this.requests.length;
      const priorityKey = Object.keys(request).find(
        (k) => typeof request[k] === "object" && "priority" in request[k]
      );
      this.requests.push({
        request,
        priority: priorityKey
          ? request[priorityKey].priority
          : DefaultQueryPriority,
        resolve: (data) => resolve(selector(data)),
        reject,
      });
    });
  }

  public async send(instanceId: string) {
    const request: QueryServiceQueryBatchBody = {
      queries: [...this.requests]
        .sort((a, b) => b.priority - a.priority)
        .map(({ request }) => request),
    };
    this.controller = new AbortController();
    const stream = streamingFetchWrapper<{ result: V1QueryBatchResponse }>(
      `http://localhost:9009/v1/instances/${instanceId}/query/batch`,
      "post",
      request,
      this.controller.signal
    );

    const hit = new Set<number>();

    for await (const res of stream) {
      const idx = res.result.key ?? 0;
      hit.add(idx);
      if (res.result.error) {
        this.requests[idx].reject(buildError(res.result.error));
        continue;
      }
      this.requests[idx].resolve(res.result);
    }

    for (let i = 0; i < this.requests.length; i++) {
      if (hit.has(i)) continue;
      this.requests[i].reject(buildError("No response"));
    }
  }

  public cancel() {
    if (!this.controller || this.controller.signal.aborted) return;
    this.controller.abort();
  }
}

function buildError(message: string): Error {
  const err = new Error(message);
  (err as any).response = {
    status: 500,
    data: {
      error: message,
    },
  };
  return err;
}
