import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";
import { copySubsetParams } from "@rilldata/web-common/lib/url-utils.ts";
import type { UrlParamsStore } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
import { page } from "$app/state";
import { goto } from "$app/navigation";

type UrlParamsChangeTrackerEvents = {
  /**
   * Fired when the underlying class's state changes.
   * Does not fire when `setUrlParams` is called to avoid loop.
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
  public searchParams = $state(new URLSearchParams());

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

    const relevantParams = copySubsetParams(urlParams, this.store.paramKeys);
    if (this.searchParams.toString() === relevantParams.toString()) return;

    this.searchParams = this.store.normalizeParams(relevantParams);
    this.store.setUrlParams(relevantParams);
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
    return this.on("change", (newUrlParams) => {
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
    if (!this.pendingParams) return;
    const pendingParams = this.pendingParams;
    // Unset before calling setUrlParams.
    // This will ensure that if params are still supposed to be pending, they are not cleared.
    this.pendingParams = undefined;

    this.setUrlParams(pendingParams);
  }

  private maybeNotifyStateChange() {
    const urlParams = new URLSearchParams();
    this.store.applyFilterToParams(urlParams);

    if (this.searchParams.toString() === urlParams.toString()) return;
    this.searchParams = urlParams;
    this.events.emit("change", this.searchParams);
  }
}
