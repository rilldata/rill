import { Heap } from "@rilldata/web-local/lib/http-request-queue/Heap";

export interface RequestQueueEntry {
  url: string;
  method: string;
  headers?: HeadersInit;
  params?: Record<string, unknown>;
  data?: any;
  signal?: AbortSignal;

  resolve?: (data: any) => void;
  reject?: (err: any) => void;
}

export interface RequestQueueQueryEntry {
  type: string;
  priority: number;
  entries: Array<RequestQueueEntry>;
}

export interface RequestQueueNameEntry {
  name: string;
  priority: number;
  queryHeap: Heap<RequestQueueQueryEntry>;
}

export type RequestQueueHeapItem =
  | RequestQueueNameEntry
  | RequestQueueQueryEntry;

const queueCompareFunction = (
  a: RequestQueueHeapItem,
  b: RequestQueueHeapItem
) => a.priority - b.priority;

export function getHeapByName(): Heap<RequestQueueNameEntry> {
  return new Heap<RequestQueueNameEntry>(
    queueCompareFunction,
    (a: RequestQueueNameEntry) => a.name
  );
}

export function getHeapByQuery(): Heap<RequestQueueQueryEntry> {
  return new Heap<RequestQueueQueryEntry>(
    queueCompareFunction,
    (a: RequestQueueQueryEntry) => a.type
  );
}
