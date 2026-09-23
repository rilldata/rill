import type { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

type UrlParamsStoreEvents = {
  ready: void;
};

export interface UrlParamsStore {
  setUrlParams(urlParams: URLSearchParams): void;
  applyFilterToParams(urlParams: URLSearchParams): void;
  ready: boolean;

  on: EventEmitter<UrlParamsStoreEvents>["on"];

  /**
   * Keys set by this class.
   */
  paramKeys: Set<string>;
  normalizeParams(urlParams: URLSearchParams): URLSearchParams;
}
