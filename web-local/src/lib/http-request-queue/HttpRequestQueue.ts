import type { RequestQueueEntry } from "@rilldata/web-local/lib/http-request-queue/HttpRequestQueueTypes";
import {
  getHeapByName,
  getHeapByQuery,
  RequestQueueNameEntry,
} from "@rilldata/web-local/lib/http-request-queue/HttpRequestQueueTypes";
import {
  ActiveColumnPriorityOffset,
  ActivePriority,
  DefaultQueryPriority,
  getPriority,
  InactivePriority,
} from "@rilldata/web-local/lib/http-request-queue/priorities";
import {
  fetchWrapper,
  FetchWrapperOptions,
} from "@rilldata/web-local/lib/util/fetchWrapper";
import { waitUntil } from "@rilldata/web-local/lib/util/waitUtils";

export const UrlExtractorRegex =
  /v1\/instances\/[\w-]*\/(metrics-views|queries)\/([\w-]*)\/([\w-]*)\/(?:([\w-]*)(?:\/|$))?/;

// intentionally 1 less than max to allow for non profiling query calls
const QueryQueueSize = 5;

/**
 * Given a URL and params this manages where the url should sit.
 * Responsible for extracting name and type and mapping to correct place.
 */
export class HttpRequestQueue {
  private readonly nameHeap = getHeapByName();
  private running = false;
  private activeCount = 0;

  public constructor(private readonly urlBase: string) {}

  public add(requestOptions: FetchWrapperOptions) {
    const urlMatch = UrlExtractorRegex.exec(requestOptions.url);
    // prepend after parsing to make parsing faster
    requestOptions.url = `${this.urlBase}${requestOptions.url}`;

    const entry: RequestQueueEntry = {
      requestOptions,
      weight: DefaultQueryPriority,
    };

    let type: string;
    let name: string;
    let columnName: string;
    let priority: number;
    switch (urlMatch?.[1]) {
      case "metrics-views":
        name = urlMatch[3];
        type = urlMatch[2];
        break;
      case "queries":
        name = urlMatch[4];
        type = urlMatch[2];
        requestOptions.params ??= {};
        priority =
          requestOptions.data?.priority ??
          (requestOptions.params.priority as number);
        columnName =
          requestOptions.params.columnName ?? requestOptions.data?.columnName;
        break;
      default:
        // make the call directly if the url is not recognised
        return fetchWrapper(requestOptions);
    }
    if (!priority) {
      priority = getPriority(type);
    }
    entry.weight = priority;
    requestOptions.params.priority = priority;

    // Adding more levels can be added here by adding more name entries under the top level one
    // Make sure to update run if so
    const nameEntry = this.getNameEntry(name);
    if (columnName) {
      entry.columnName = columnName;
      entry.key = `${columnName}-${type}`;
      if (!nameEntry.columnMap.has(columnName)) {
        nameEntry.columnMap.set(columnName, []);
      }
      nameEntry.columnMap.get(columnName).push(entry);
    } else {
      entry.key = type;
    }
    nameEntry.queryHeap.push(entry);
    this.run();

    return new Promise((resolve, reject) => {
      entry.resolve = resolve;
      entry.reject = reject;
    });
  }

  public removeByName(name: string) {
    this.nameHeap.delete(undefined, name);
  }

  public inactiveByName(name: string) {
    const nameEntry = this.nameHeap.get(name);
    if (!nameEntry) return;
    nameEntry.weight = InactivePriority;
    this.nameHeap.updateItem(nameEntry);
  }

  public prioritiseColumn(name: string, columnName: string, increase: boolean) {
    const nameEntry = this.nameHeap.get(name);
    if (!nameEntry) return;
    const columnEntries = nameEntry.columnMap.get(columnName);
    if (!columnEntries) return;
    columnEntries.forEach((columnEntry) => {
      if (increase && columnEntry.weight < ActiveColumnPriorityOffset) {
        columnEntry.weight += ActiveColumnPriorityOffset;
      } else if (columnEntry.weight > ActiveColumnPriorityOffset) {
        columnEntry.weight -= ActiveColumnPriorityOffset;
      }
      nameEntry.queryHeap.updateItem(columnEntry);
    });
  }

  private async run() {
    if (this.running) return;
    this.running = true;

    while (!this.nameHeap.empty()) {
      await waitUntil(() => this.activeCount < QueryQueueSize, 30000, 50);
      if (this.activeCount >= QueryQueueSize) continue;
      if (this.nameHeap.empty()) break;

      const topNameEntry = this.nameHeap.peek();
      const entry = topNameEntry.queryHeap.pop();

      // intentional to not await here
      this.fireForEntry(entry);
      if (entry.columnName && topNameEntry.columnMap.has(entry.columnName)) {
        this.clearEntryForColumn(topNameEntry, entry);
      }

      // cleanup
      if (topNameEntry.queryHeap.empty()) {
        this.nameHeap.pop();
      }
    }

    this.running = false;
  }

  private getNameEntry(name: string): RequestQueueNameEntry {
    let nameEntry = this.nameHeap.get(name);
    if (!nameEntry) {
      nameEntry = {
        name,
        weight: ActivePriority,
        columnMap: new Map(),
        queryHeap: getHeapByQuery(),
      };
      this.nameHeap.push(nameEntry);
    }
    return nameEntry;
  }

  private async fireForEntry(entry: RequestQueueEntry) {
    this.activeCount++;
    try {
      const resp = await fetchWrapper(entry.requestOptions);
      entry.resolve(resp);
    } catch (err) {
      entry.reject(err);
    }
    this.activeCount--;
  }

  private clearEntryForColumn(
    nameEntry: RequestQueueNameEntry,
    entry: RequestQueueEntry
  ) {
    const entriesForColumn = nameEntry.columnMap.get(entry.columnName);
    const index = entriesForColumn.indexOf(entry);
    if (index === -1) return;
    entriesForColumn.splice(index, 1);
  }
}
