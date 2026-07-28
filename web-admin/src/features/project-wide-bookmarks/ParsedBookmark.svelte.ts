import type { V1Bookmark } from "@rilldata/web-admin/client";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { page } from "$app/state";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { ExploreSpecProvider } from "@rilldata/web-common/features/dashboards/ExploreSpecProvider.svelte.ts";

interface BookmarkParser {
  // Url params directly converted from bookmark.
  // Default url params are cleaned to match the actual url.
  bookmarkUrlParams: URLSearchParams;

  cleanup: (() => void) | undefined;
}

export class DefaultBookmarkParser implements BookmarkParser {
  public bookmarkUrlParams: URLSearchParams;
  public cleanup: (() => void) | undefined = undefined;

  public constructor(bookmark: V1Bookmark) {
    this.bookmarkUrlParams = $state(
      new URLSearchParams(bookmark.urlSearch ?? ""),
    );
  }
}

// TODO: since we have a provider this can be moved directly to ParsedBookmark
export class ExploreBookmarkParser implements BookmarkParser {
  public bookmarkUrlParams: URLSearchParams;
  public cleanup: (() => void) | undefined = undefined;

  public constructor(runtimeClient: RuntimeClient, bookmark: V1Bookmark) {
    const provider = new ExploreSpecProvider(
      runtimeClient,
      bookmark.resourceName ?? "",
    );

    this.cleanup = provider.cleanup;

    this.bookmarkUrlParams = $derived.by(() => {
      let searchParams: URLSearchParams;
      if (bookmark.data) {
        const exploreStateFromBookmark = getDashboardStateFromUrl(
          bookmark.data,
          provider.metricsViewSpec,
          provider.exploreSpec,
        );

        // We need to check if the bookmark's url is equal to current url or not to show an "active" state.
        // To avoid calculating it everytime we directly convert it to final url.
        searchParams = convertPartialExploreStateToUrlParams(
          provider.exploreSpec,
          provider.metricsViewSpec,
          exploreStateFromBookmark,
          getTimeControlState(
            provider.metricsViewSpec,
            provider.exploreSpec,
            provider.timeRangeSummary,
            exploreStateFromBookmark,
          ),
        );
      } else {
        searchParams = new URLSearchParams(bookmark.urlSearch ?? "");
      }

      const defaultUrlParams = getRillDefaultExploreUrlParams(
        provider.metricsViewSpec,
        provider.exploreSpec,
        provider.timeRangeSummary,
      );
      return cleanUrlParams(searchParams, defaultUrlParams);
    });
  }
}

export class ParsedBookmark {
  // If bookmark has filter-only params then this is true. Non-filter params from the current url are retained.
  public filtersOnly: boolean;
  // If bookmark has an absolute time range. Start and end times will be hardcoded here.
  public absoluteTimeRange: boolean;
  public isActive: boolean;
  // Full url to be used to navigate.
  // This contains existing non-filter params on top of filter params for filters only bookmark.
  public fullUrl: string;

  public cleanup: (() => void) | undefined;

  public constructor(
    runtimeClient: RuntimeClient,
    public bookmark: V1Bookmark,
  ) {
    const parser =
      bookmark.resourceKind === ResourceKind.Explore
        ? new ExploreBookmarkParser(runtimeClient, bookmark)
        : new DefaultBookmarkParser(bookmark);

    this.filtersOnly = $derived(isFilterOnlyBookmark(parser.bookmarkUrlParams));
    this.absoluteTimeRange = $derived(
      isAbsoluteTimeRangeBookmark(parser.bookmarkUrlParams),
    );

    this.isActive = $derived.by(() => {
      if (!this.filtersOnly) {
        return (
          parser.bookmarkUrlParams.toString() ===
          page.url.searchParams.toString()
        );
      }

      return [...parser.bookmarkUrlParams.entries()].every(([key, value]) => {
        const curValue = page.url.searchParams.get(key);
        return curValue === value;
      });
    });

    this.fullUrl = $derived.by(() => {
      const fullUrlParams = new URLSearchParams(parser.bookmarkUrlParams);
      if (this.filtersOnly) {
        page.url.searchParams.forEach((v, p) => {
          if (fullUrlParams.has(p)) return;
          fullUrlParams.set(p, v);
        });
      }

      const projectPath = `${page.params.organization}/${page.params.project}`;
      const resourceRoute =
        this.bookmark.resourceKind === ResourceKind.Explore
          ? "explore"
          : "canvas";
      const dashboardPath = `${resourceRoute}/${this.bookmark.resourceName ?? ""}`;

      return `/${projectPath}/${dashboardPath}?${fullUrlParams.toString()}`;
    });
  }
}

// These are the only parameters that are stored in a filter-only bookmark
const FILTER_ONLY_PARAMS = new Set([
  ExploreStateURLParams.Filters,
  ExploreStateURLParams.TimeRange,
  ExploreStateURLParams.TimeGrain,
]) as Set<string>;

function isFilterOnlyBookmark(bookmarkUrlParams: URLSearchParams): boolean {
  const urlParams = Array.from(bookmarkUrlParams.keys());
  return urlParams.every((param) => FILTER_ONLY_PARAMS.has(param));
}

function isAbsoluteTimeRangeBookmark(bookmarkUrlParams: URLSearchParams) {
  const timeRange = bookmarkUrlParams.get(ExploreStateURLParams.TimeRange);
  if (!timeRange) return false;

  try {
    const rt = parseRillTime(timeRange);
    return rt.isAbsoluteTime();
  } catch {
    return false;
  }
}
