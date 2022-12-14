import type { RequestQueueEntry } from "@rilldata/web-local/lib/http-request-queue/HttpRequestQueueTypes";
import {
  getHeapByName,
  getHeapByQuery,
  RequestQueueNameEntry,
  RequestQueueQueryEntry,
} from "@rilldata/web-local/lib/http-request-queue/HttpRequestQueueTypes";
import {
  ActivePriority,
  DefaultQueryPriority,
  InactivePriority,
  QueryPriorities,
} from "@rilldata/web-local/lib/http-request-queue/priorities";
import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";
import { waitUntil } from "@rilldata/web-local/lib/util/waitUtils";

// TODO: timeseries
const UrlExtractorRegex =
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

  public add(entry: RequestQueueEntry) {
    const urlMatch = UrlExtractorRegex.exec(entry.url);
    // prepend after parsing to make parsing faster
    entry.url = `${this.urlBase}${entry.url}`;

    let type: string;
    let name: string;
    switch (urlMatch?.[1]) {
      case "metrics-views":
        name = urlMatch[3];
        type = urlMatch[2];
        break;
      case "queries":
        name = urlMatch[4];
        type = urlMatch[2];
        entry.params ??= {};
        entry.params.priority = QueryPriorities[type] ?? DefaultQueryPriority;
        if (entry.data) {
          entry.data.priority = entry.params.priority;
        }
        break;
      default:
        // make the call directly if the url is not recognised
        return fetchWrapper(entry);
    }

    // Adding more levels can be added here by adding more name entries under the top level one
    // Make sure to update run if so
    const nameEntry = this.getNameEntry(name);
    const typeEntry = this.getTypeEntry(nameEntry, type);
    typeEntry.entries.push(entry);
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
    nameEntry.priority = InactivePriority;
    this.nameHeap.updateItem(nameEntry);
  }

  private async run() {
    if (this.running) return;
    this.running = true;

    while (!this.nameHeap.empty()) {
      await waitUntil(() => this.activeCount < QueryQueueSize, 30000, 50);
      if (this.activeCount >= QueryQueueSize) continue;
      if (this.nameHeap.empty()) break;

      const topNameEntry = this.nameHeap.peek();
      const topTypeEntry = topNameEntry.queryHeap.peek();

      // intentional to not await here
      this.fireForEntry(topTypeEntry.entries.shift());

      // cleanup
      if (topTypeEntry.entries.length === 0) {
        topNameEntry.queryHeap.pop();

        if (topNameEntry.queryHeap.empty()) {
          this.nameHeap.pop();
        }
      }
    }

    this.running = false;
  }

  private getNameEntry(name: string): RequestQueueNameEntry {
    let nameEntry = this.nameHeap.get(name);
    if (!nameEntry) {
      nameEntry = {
        name,
        priority: ActivePriority,
        queryHeap: getHeapByQuery(),
      };
      this.nameHeap.push(nameEntry);
    }
    return nameEntry;
  }

  private getTypeEntry(
    nameEntry: RequestQueueNameEntry,
    type: string
  ): RequestQueueQueryEntry {
    let typeEntry = nameEntry.queryHeap.get(type);
    if (!typeEntry) {
      typeEntry = {
        type,
        priority: QueryPriorities[type] ?? DefaultQueryPriority,
        entries: [],
      };
      nameEntry.queryHeap.push(typeEntry);
    }
    return typeEntry;
  }

  private async fireForEntry(entry: RequestQueueEntry) {
    this.activeCount++;
    try {
      const resp = await fetchWrapper(entry);
      entry.resolve(resp);
    } catch (err) {
      entry.reject(err);
    }
    this.activeCount--;
  }
}
