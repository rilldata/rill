import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/fetch-streaming-wrapper";
import {
  QueryServiceQueryBatchBody,
  V1Query,
  V1QueryBatchResponse,
  V1QueryResult,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get, Readable, writable } from "svelte/store";

export type StreamingQueryBatchResponse<Resp> = {
  responses: Resp[];
  errors: Error[];
  completed: number;
  progress: number;
};

export function createQueryServiceStreamingQueryBatch<Resp = V1QueryResult>(
  instanceId: string,
  queries: V1Query[],
  select: (resp: V1QueryResult, index: number) => Resp,
): Readable<StreamingQueryBatchResponse<Resp>> {
  const { subscribe, update } = writable<StreamingQueryBatchResponse<Resp>>({
    responses: new Array<Resp>(queries.length),
    errors: [],
    completed: 0,
    progress: 0,
  });

  const handleResp = (resp: V1QueryResult, index: number) => {
    update((s) => {
      s.responses[index] = select(resp, index);
      s.completed++;
      s.progress = Math.round((s.completed * 100) / queries.length);
      return s;
    });
  };
  const handleError = (err: Error) => {
    update((s) => {
      s.errors.push(err);
      s.completed++;
      s.progress = Math.round((s.completed * 100) / queries.length);
      return s;
    });
  };

  queryServiceStreamingQueryBatch(
    instanceId,
    queries.map((q) => [q, undefined, handleResp, handleError]),
  ).catch((err) => {
    update((s) => {
      s.errors.unshift(err instanceof Error ? err : new Error(err));
      s.completed = queries.length;
      s.progress = 100;
      return s;
    });
  });

  return {
    subscribe,
  };
}

export async function queryServiceStreamingQueryBatch(
  instanceId: string,
  queries: [
    query: V1Query,
    abortSignal: AbortSignal | undefined,
    resolve: (resp: V1QueryResult, index: number) => void,
    reject: (err: Error) => void,
  ][],
) {
  const headers = { "Content-Type": "application/json" };
  const jwt = get(runtime).jwt;
  if (jwt) {
    headers["Authorization"] = `Bearer ${jwt.token}`;
  }

  const body: QueryServiceQueryBatchBody = {
    queries: [],
  };
  const controller = new AbortController();
  queries.forEach(([query, signal]) => {
    body.queries?.push(query);
    signal?.addEventListener("abort", () => {
      controller.abort();
    });
  });

  const stream = streamingFetchWrapper<{ result: V1QueryBatchResponse }>(
    `${get(runtime).host}/v1/instances/${instanceId}/query/batch`,
    "POST",
    body,
    headers,
    controller.signal,
  );

  const seen = new Set<number>();
  for await (const { result } of stream) {
    if (result.index === undefined || !queries[result.index]) continue;
    seen.add(result.index);
    if (result.result) {
      queries[result.index][2](result.result, result.index);
    } else {
      queries[result.index][3](new Error(result.error ?? "no response"));
    }
  }

  queries.forEach(([, , , reject], i) => {
    if (seen.has(i)) return;
    reject(new Error("no response"));
  });
}
