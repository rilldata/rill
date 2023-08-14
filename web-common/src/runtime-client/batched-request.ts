import type {
  QueryServiceQueryBatchBody,
  V1QueryBatchEntry,
  V1QueryBatchResponse,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/streaming-fetch-wrapper";

export type BatchRequest = {
  request: V1QueryBatchEntry;
  priority: number;
  resolve: (data: V1QueryBatchResponse) => void;
  reject: (err: Error) => void;
  signal: AbortSignal | undefined;
};
export class BatchedRequest {
  private requests = new Array<BatchRequest>();

  public add(
    request: V1QueryBatchEntry,
    priority: number,
    resolve: (data: V1QueryBatchResponse) => void,
    reject: () => void,
    signal: AbortSignal | undefined
  ) {
    request.key = this.requests.length;
    this.requests.push({ request, priority, resolve, reject, signal });
  }

  public async send(instanceId: string) {
    const request: QueryServiceQueryBatchBody = {
      queries: [...this.requests]
        .sort((e1, e2) => e2.priority - e1.priority)
        .map(({ request }) => request),
    };
    const controller = new AbortController();
    const stream = streamingFetchWrapper<{ result: V1QueryBatchResponse }>(
      `http://localhost:9009/v1/instances/${instanceId}/query/batch`,
      "post",
      request,
      controller.signal
    );

    this.requests.forEach(({ signal }) => {
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
