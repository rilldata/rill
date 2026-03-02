import type { Interceptor } from "@connectrpc/connect";
import { appQueryStatusStore } from "../application-store";
import { Heap } from "./heap";
import {
  ACTIVE_COLUMN_PRIORITY_OFFSET,
  ACTIVE_PRIORITY,
  INACTIVE_PRIORITY,
  getPriorityForMethod,
} from "./request-priorities";

interface QueueEntry {
  key: string;
  weight: number;
  id: number;
  columnName?: string;

  resolve?: (value: unknown) => void;
  reject?: (reason: unknown) => void;
  fn?: () => Promise<unknown>;

  index?: number;
}

interface NameEntry {
  name: string;
  weight: number;
  columnMap: Map<string, Array<QueueEntry>>;
  queryHeap: Heap<QueueEntry>;

  index?: number;
}

function createNameHeap(): Heap<NameEntry> {
  return new Heap<NameEntry>(
    (a, b) => a.weight - b.weight,
    (a) => a.name,
  );
}

function createQueryHeap(): Heap<QueueEntry> {
  return new Heap<QueueEntry>(
    (a, b) => {
      if (a.weight === b.weight) {
        // equal weights: preserve insertion order (lower id = earlier)
        return b.id - a.id;
      }
      return a.weight - b.weight;
    },
    (a) => a.key,
  );
}

function getDefaultConcurrency(): number {
  try {
    if (
      window.location.protocol === "https:" ||
      window.location.hostname !== "localhost"
    ) {
      return 200;
    }
  } catch {
    // no-op: SSR or test environment
  }
  return 5;
}

export class RequestQueue {
  private readonly nameHeap = createNameHeap();
  private activeCount = 0;
  private ids = 0;
  private readonly maxConcurrent: number;

  constructor(opts?: { maxConcurrent?: number }) {
    this.maxConcurrent = opts?.maxConcurrent ?? getDefaultConcurrency();
  }

  /**
   * Queue a request. Resolves when the request function completes.
   * The request won't fire until the queue has capacity.
   */
  enqueue<T>(opts: {
    priority: number;
    resourceName?: string;
    columnName?: string;
    signal?: AbortSignal;
    fn: () => Promise<T>;
  }): Promise<T> {
    const resourceName = opts.resourceName ?? "__default__";
    const nameEntry = this.getNameEntry(resourceName);

    const entry: QueueEntry = {
      key: `${opts.columnName ?? ""}:${this.ids}`,
      weight: opts.priority,
      id: this.ids++,
      columnName: opts.columnName,
      fn: opts.fn as () => Promise<unknown>,
    };

    if (opts.columnName) {
      if (!nameEntry.columnMap.has(opts.columnName)) {
        nameEntry.columnMap.set(opts.columnName, []);
      }
      nameEntry.columnMap.get(opts.columnName)!.push(entry);
    }

    nameEntry.queryHeap.push(entry);

    const promise = new Promise<T>((resolve, reject) => {
      entry.resolve = resolve as (value: unknown) => void;
      entry.reject = reject;
    });

    // If already aborted, reject immediately
    if (opts.signal?.aborted) {
      nameEntry.queryHeap.delete(entry);
      if (entry.columnName) {
        this.clearEntryForColumn(nameEntry, entry);
      }
      if (nameEntry.queryHeap.empty()) {
        this.nameHeap.delete(nameEntry);
      }
      entry.reject?.(opts.signal.reason);
      return promise;
    }

    // Handle abort while queued
    opts.signal?.addEventListener(
      "abort",
      () => {
        // Only remove if still queued (has fn; cleared on fire)
        if (!entry.fn) return;
        nameEntry.queryHeap.delete(entry);
        if (entry.columnName) {
          this.clearEntryForColumn(nameEntry, entry);
        }
        if (nameEntry.queryHeap.empty()) {
          this.nameHeap.delete(nameEntry);
        }
        entry.reject?.(opts.signal!.reason);
      },
      { once: true },
    );

    this.drain();
    return promise;
  }

  /** Boost or reduce priority for column-level queries (active column in Explore). */
  prioritiseColumn(
    resourceName: string,
    columnName: string,
    active: boolean,
  ): void {
    const nameEntry = this.nameHeap.get(resourceName);
    if (!nameEntry) return;
    const columnEntries = nameEntry.columnMap.get(columnName);
    if (!columnEntries) return;
    for (const entry of columnEntries) {
      if (active && entry.weight < ACTIVE_COLUMN_PRIORITY_OFFSET) {
        entry.weight += ACTIVE_COLUMN_PRIORITY_OFFSET;
      } else if (!active && entry.weight > ACTIVE_COLUMN_PRIORITY_OFFSET) {
        entry.weight -= ACTIVE_COLUMN_PRIORITY_OFFSET;
      }
      nameEntry.queryHeap.updateItem(entry);
    }
  }

  /** Remove all queued requests for a resource (entity deleted/renamed). */
  removeByName(resourceName: string): void {
    const nameEntry = this.nameHeap.get(resourceName);
    if (!nameEntry) return;
    while (!nameEntry.queryHeap.empty()) {
      const entry = nameEntry.queryHeap.pop()!;
      entry.reject?.(new DOMException("Request cancelled", "AbortError"));
    }
    this.nameHeap.delete(nameEntry);
  }

  /** Reject all pending requests (client is being disposed). */
  clear(): void {
    while (!this.nameHeap.empty()) {
      const nameEntry = this.nameHeap.pop()!;
      while (!nameEntry.queryHeap.empty()) {
        const entry = nameEntry.queryHeap.pop()!;
        entry.reject?.(new DOMException("Request cancelled", "AbortError"));
      }
    }
  }

  /** Deprioritize a resource (user navigated away). */
  inactiveByName(resourceName: string): void {
    const nameEntry = this.nameHeap.get(resourceName);
    if (!nameEntry) return;
    nameEntry.weight = INACTIVE_PRIORITY;
    this.nameHeap.updateItem(nameEntry);
  }

  private drain(): void {
    while (!this.nameHeap.empty() && this.activeCount < this.maxConcurrent) {
      const topNameEntry = this.nameHeap.peek();
      const entry = topNameEntry.queryHeap.pop();
      if (!entry) break;

      this.fireEntry(entry);
      this.activeCount++;

      if (entry.columnName && topNameEntry.columnMap.has(entry.columnName)) {
        this.clearEntryForColumn(topNameEntry, entry);
      }

      if (topNameEntry.queryHeap.empty()) {
        this.nameHeap.pop();
      }
    }
    appQueryStatusStore.set(this.activeCount > 0);
  }

  private async fireEntry(entry: QueueEntry): Promise<void> {
    const fn = entry.fn!;
    entry.fn = undefined; // mark as fired
    try {
      const result = await fn();
      entry.resolve?.(result);
    } catch (err) {
      entry.reject?.(err);
    }
    this.activeCount--;
    appQueryStatusStore.set(this.activeCount > 0);
    this.drain();
  }

  private getNameEntry(name: string): NameEntry {
    let nameEntry = this.nameHeap.get(name);
    if (!nameEntry) {
      nameEntry = {
        name,
        weight: ACTIVE_PRIORITY,
        columnMap: new Map(),
        queryHeap: createQueryHeap(),
      };
      this.nameHeap.push(nameEntry);
    }
    return nameEntry;
  }

  private clearEntryForColumn(nameEntry: NameEntry, entry: QueueEntry): void {
    const entries = nameEntry.columnMap.get(entry.columnName!);
    if (!entries) return;
    const idx = entries.indexOf(entry);
    if (idx !== -1) entries.splice(idx, 1);
  }
}

/**
 * Extract the resource name from a ConnectRPC request message.
 * Looks for common field patterns in runtime API request messages.
 */
function extractResourceName(req: { message: unknown }): string | undefined {
  const msg = req.message as Record<string, unknown>;
  // MetricsView queries use metricsViewName or metricsView
  if (typeof msg.metricsViewName === "string") return msg.metricsViewName;
  if (typeof msg.metricsView === "string") return msg.metricsView;
  // Column profiling queries use tableName
  if (typeof msg.tableName === "string") return msg.tableName;
  return undefined;
}

/**
 * Creates a ConnectRPC interceptor that routes requests through the RequestQueue.
 * The interceptor wraps the `next()` call so the queue controls when requests fire.
 */
export function createQueueInterceptor(queue: RequestQueue): Interceptor {
  return (next) => async (req) => {
    const priority = getPriorityForMethod(req.method.name);
    const resourceName = extractResourceName(req);
    const columnName = (req.message as Record<string, unknown>).columnName as
      | string
      | undefined;

    return queue.enqueue({
      priority,
      resourceName,
      columnName,
      signal: req.signal,
      fn: () => next(req),
    });
  };
}
