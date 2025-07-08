import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
import {
  fetchWrapper,
  type FetchWrapperOptions,
} from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { RequestQueueEntry } from "@rilldata/web-common/runtime-client/http-request-queue/HttpRequestQueueTypes";
import {
  getHeapByName,
  getHeapByQuery,
  type RequestQueueNameEntry,
} from "@rilldata/web-common/runtime-client/http-request-queue/HttpRequestQueueTypes";
import {
  ActiveColumnPriorityOffset,
  ActivePriority,
  DefaultQueryPriority,
  getPriority,
  InactivePriority,
} from "@rilldata/web-common/runtime-client/http-request-queue/priorities";

// Examples:
// v1/instances/id/queries/columns-profile/tables/table-name
// v1/instances/id/queries/metrics-views/mv-name/timeseries
export const UrlExtractorRegex =
  /v1\/instances\/[\w-]+\/queries\/([\w-]+)\/([\w-]+)\/([\w-]+)/;

// intentionally 1 less than max to allow for non profiling query calls
let QueryQueueSize = 5;
try {
  if (
    window.location.protocol === "https:" ||
    window.location.hostname !== "localhost"
  ) {
    QueryQueueSize = 200;
  }
} catch {
  // no-op
}

/**
 * Given a URL and params this manages where the url should sit.
 * Responsible for extracting name and type and mapping to correct place.
 */
export class HttpRequestQueue {
  private readonly nameHeap = getHeapByName();
  private activeCount = 0;
  private ids = 0;

  public constructor() {
    // no-op
  }

  public add(requestOptions: FetchWrapperOptions) {
    // prepend after parsing to make parsing faster
    requestOptions.url = `${requestOptions?.baseUrl}${requestOptions.url}`;

    const urlMatch = UrlExtractorRegex.exec(requestOptions.url);

    let type: string;
    let name: string;
    if (urlMatch) {
      if (urlMatch[1] === "metrics-views") {
        name = urlMatch[2];
        type = urlMatch[3];
      } else {
        name = urlMatch[3];
        type = urlMatch[1];
      }
    } else {
      // make the call directly if the url is not recognised
      return fetchWrapper(requestOptions);
    }

    const entry: RequestQueueEntry = {
      requestOptions,
      weight: DefaultQueryPriority,
    };

    let priority: number;
    requestOptions.params ??= {};
    priority =
      requestOptions.data?.priority ??
      (requestOptions.params.priority as number);
    const columnName =
      requestOptions.params.columnName ??
      requestOptions.data?.columnName ??
      requestOptions.data?.dimensionName;

    if (!priority) {
      priority = getPriority(type);
    }
    // Skip adding priority parameter for canvas resolve requests
    if (type !== "canvases") {
      requestOptions.params.priority = priority;
    }
    entry.weight = priority;
    entry.id = this.ids++;

    // Adding more levels can be added here by adding more name entries under the top level one
    // Make sure to update popEntries if so
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
    // intentional to not await here
    this.popEntries();

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

  private async popEntries() {
    while (!this.nameHeap.empty() && this.activeCount < QueryQueueSize) {
      const topNameEntry = this.nameHeap.peek();
      const entry = topNameEntry.queryHeap.pop();

      // intentional to not await here
      this.fireForEntry(entry);
      this.activeCount++;
      if (entry.columnName && topNameEntry.columnMap.has(entry.columnName)) {
        this.clearEntryForColumn(topNameEntry, entry);
      }

      // cleanup
      if (topNameEntry.queryHeap.empty()) {
        this.nameHeap.pop();
      }
    }
    appQueryStatusStore.set(this.activeCount > 0);
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
    try {
      const resp = await fetchWrapper(entry.requestOptions);
      if (entry?.resolve) entry.resolve(resp);
    } catch (err) {
      if (entry?.reject) entry.reject(err);
    }
    this.activeCount--;
    appQueryStatusStore.set(this.activeCount > 0);
    return this.popEntries();
  }

  private clearEntryForColumn(
    nameEntry: RequestQueueNameEntry,
    entry: RequestQueueEntry,
  ) {
    const entriesForColumn = nameEntry.columnMap.get(entry.columnName);
    const index = entriesForColumn.indexOf(entry);
    if (index === -1) return;
    entriesForColumn.splice(index, 1);
  }
}
