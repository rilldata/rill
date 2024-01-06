import type { FetchWrapperOptions } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { Heap } from "@rilldata/web-common/runtime-client/http-request-queue/Heap";

export interface RequestQueueEntry {
  requestOptions: FetchWrapperOptions;
  weight: number;

  key?: string;
  id?: number;
  columnName?: string;

  resolve?: (data: any) => void;
  reject?: (err: any) => void;

  index?: number;
}

export interface RequestQueueNameEntry {
  name: string;
  weight: number;
  columnMap: Map<string, Array<RequestQueueEntry>>;
  queryHeap: Heap<RequestQueueEntry>;

  index?: number;
}

export type RequestQueueHeapItem = RequestQueueNameEntry | RequestQueueEntry;

export function getHeapByName(): Heap<RequestQueueNameEntry> {
  return new Heap<RequestQueueNameEntry>(
    (a, b) => a.weight - b.weight,
    (a) => a.name,
  );
}

export function getHeapByQuery(): Heap<RequestQueueEntry> {
  return new Heap<RequestQueueEntry>(
    (a, b) => {
      if (a.weight === b.weight) {
        // if weights are same use the auto-incremented id.
        // this will keep the insert order
        return b.id - a.id;
      }
      return a.weight - b.weight;
    },
    (a) => a.key,
  );
}
