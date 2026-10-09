import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";
import { expandCompressedParams } from "@rilldata/web-common/features/dashboards/url-state/compression.ts";
import { copySubsetParams } from "@rilldata/web-common/lib/url-utils.ts";
import { goto } from "$app/navigation";
import { page } from "$app/state";

type UrlParamsStoreEvents = {
  ready: void;
};

export interface UrlParamsStore {
  setUrlParams(urlParams: URLSearchParams): void;
  applyFilterToParams(urlParams: URLSearchParams): void;
  ready: boolean;

  on: EventEmitter<UrlParamsStoreEvents>["on"];

  /**
   * Url param keys set by this class.
   */
  paramKeys: Set<string>;
  normalizeParams(urlParams: URLSearchParams): URLSearchParams;
}

type UrlParamsChangeTrackerEvents = {
  /**
   * Fired when the underlying class's state changes internally.
   * Happens when UI controls directly update class's state.
   * Does not fire when `setUrlParams` is called to avoid loop.
   */
  "internal-change": URLSearchParams;
  /**
   * Fired when
   */
  change: URLSearchParams;
};

/**
 * A class that tracks URL search params that splits into multiple stores but backed by a single class.
 *
 * External sources like navigation should call `setUrlParams` and listen to `change` to sync.
 * The internal store class that expands the search params into multiple stores should call `stateChanged` and listens to `set` to sync.
 */
export class UrlParamsChangeTracker {
  /**
   * Source of truth for the underlying class's state.
   */
  public searchParams = $state<URLSearchParams | undefined>();

  private pendingParams: URLSearchParams | undefined = undefined;

  private events = new EventEmitter<UrlParamsChangeTrackerEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  public constructor(private readonly store: UrlParamsStore) {
    this.store.on("ready", () => this.replayPendingParams());
  }

  /**
   * Called by external sources like navigation to sync the search params.
   * Calls `setUrlParams` on the store class.
   */
  public setUrlParams(urlParams: URLSearchParams) {
    if (!this.store.ready) {
      this.pendingParams = urlParams;
      return;
    }

    let expandedUrlParams: URLSearchParams;
    try {
      expandedUrlParams = expandCompressedParams(urlParams);
    } catch {
      // If we fail to decompress, do not throw here.
      return;
    }

    const relevantParams = copySubsetParams(
      expandedUrlParams,
      this.store.paramKeys,
    );
    if (this.searchParams?.toString() === relevantParams.toString()) return;

    this.searchParams = this.store.normalizeParams(relevantParams);
    this.store.setUrlParams(this.searchParams);
    this.events.emit("change", this.searchParams);
  }

  /**
   * Called by internal store class when its state is directly changed like a user action.
   * Calls `applyFilterToParams` to get the actual url params in a new microtask so that changes are propagated.
   * Fires `change` event so that external sources like navigation can sync.
   */
  public stateChanged() {
    queueMicrotask(() => this.maybeNotifyStateChange());
  }

  /**
   * Syncs the store's state to the URL by listening to 'change' event.
   * Returns the unsub method from the event listener. It is the callers' responsibility to call it.
   * @param emptySearchOverride The search override to use if the store's state is empty.
   */
  public syncToUrl(emptySearchOverride = "") {
    return this.on("internal-change", (newUrlParams) => {
      const urlParamsToApply = new URLSearchParams(page.url.searchParams);
      this.store.paramKeys.forEach((key) => {
        if (newUrlParams.has(key)) {
          urlParamsToApply.set(key, newUrlParams.get(key) ?? "");
        } else {
          urlParamsToApply.delete(key);
        }
      });

      let searchToApply = urlParamsToApply.toString();
      if (!searchToApply) searchToApply = emptySearchOverride;
      void goto("?" + searchToApply);
    });
  }

  private replayPendingParams() {
    if (!this.pendingParams) {
      this.reapplyCurrentParams();
      return;
    }
    const pendingParams = this.pendingParams;
    // Unset before calling setUrlParams.
    // This will ensure that if params are still supposed to be pending, they are not cleared.
    this.pendingParams = undefined;

    this.setUrlParams(pendingParams);
  }

  /**
   * Parses the current params again once the store is ready after a change in its dependencies.
   * Goes around `setUrlParams` since the params themselves have not changed.
   * Parts the store can no longer represent, like a filter on a dropped dimension, are removed,
   * and `stateChanged` reports the removal.
   */
  private reapplyCurrentParams() {
    if (!this.searchParams) return;

    this.searchParams = this.store.normalizeParams(
      copySubsetParams(this.searchParams, this.store.paramKeys),
    );
    this.store.setUrlParams(this.searchParams);
    this.stateChanged();
  }

  private maybeNotifyStateChange() {
    const urlParams = new URLSearchParams();
    this.store.applyFilterToParams(urlParams);

    if (this.searchParams?.toString() === urlParams.toString()) return;
    this.searchParams = urlParams;
    this.events.emit("internal-change", this.searchParams);
    this.events.emit("change", this.searchParams);
  }
}
