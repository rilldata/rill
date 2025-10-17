import type { V1Bookmark } from "@rilldata/web-admin/client";
import { isHomeBookmark } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { type DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";

export type BookmarkEntry = {
  resource: V1Bookmark;
  filtersOnly: boolean;
  absoluteTimeRange: boolean;
  url: string;
};

export type Bookmarks = {
  home: BookmarkEntry | undefined;
  personal: BookmarkEntry[];
  shared: BookmarkEntry[];
};

export type BookmarkFormValues = {
  displayName: string;
  description: string;
  shared: "false" | "true";
  filtersOnly: boolean;
  absoluteTimeRange: boolean;
};

// These are the only parameters that are stored in a filter-only bookmark
const FILTER_ONLY_PARAMS = new Set([
  ExploreStateURLParams.Filters,
  ExploreStateURLParams.TimeRange,
  ExploreStateURLParams.TimeGrain,
]) as Set<string>;

export function getBookmarkData({
  curUrlParams,
  defaultUrlParams,
  filtersOnly,
  absoluteTimeRange,
  selectedTimeRange,
  selectedComparisonTimeRange,
}: {
  curUrlParams: URLSearchParams;
  defaultUrlParams?: URLSearchParams;
  filtersOnly?: boolean;
  absoluteTimeRange?: boolean;
  selectedTimeRange?: DashboardTimeControls;
  selectedComparisonTimeRange?: DashboardTimeControls;
}) {
  // Create a copy to avoid mutating the source.
  const bookmarkUrlParams = new URLSearchParams(curUrlParams);

  // Merge defaults that are not present in the source. This is mandatory in explore where default params are ommitted.
  if (defaultUrlParams) {
    for (const [key, value] of defaultUrlParams) {
      if (!bookmarkUrlParams.has(key)) {
        bookmarkUrlParams.set(key, value);
      }
    }
  }

  // If the bookmark is for absolute time range. Update the time range and compare time range params with resolved start and end.
  if (absoluteTimeRange && selectedTimeRange.start && selectedTimeRange.end) {
    bookmarkUrlParams.set(
      ExploreStateURLParams.TimeRange,
      `${selectedTimeRange.start.toISOString()},${selectedTimeRange.end.toISOString()}`,
    );

    if (selectedComparisonTimeRange.start && selectedComparisonTimeRange.end) {
      bookmarkUrlParams.set(
        ExploreStateURLParams.ComparisonTimeRange,
        `${selectedComparisonTimeRange.start.toISOString()},${selectedComparisonTimeRange.end.toISOString()}`,
      );
    }
  }

  // If the bookmark is filter only, then only keep retain FILTER_ONLY_PARAMS from bookmarkUrlParams.
  if (filtersOnly) {
    const filterOnlyUrlParams = new URLSearchParams();
    FILTER_ONLY_PARAMS.forEach((param) => {
      if (bookmarkUrlParams.has(param)) {
        filterOnlyUrlParams.set(param, bookmarkUrlParams.get(param));
      }
    });

    return "?" + filterOnlyUrlParams.toString();
  }

  return "?" + bookmarkUrlParams.toString();
}

export function formatTimeRange(
  start: string,
  end: string,
  timeGrain: V1TimeGrain | undefined,
  timezone: string | undefined,
) {
  timezone ??= "UTC";
  timeGrain ??= V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  const startTime = DateTime.fromISO(start).setZone(timezone);
  const endTime = DateTime.fromISO(end).setZone(timezone);
  const interval = Interval.fromDateTimes(startTime, endTime);
  if (!interval.isValid) return "";
  return prettyFormatTimeRange(interval, timeGrain);
}

export function parseBookmarks(
  bookmarkResp: V1Bookmark[],
  dataTransformer: (data: string) => string = (data) => data,
) {
  return bookmarkResp.map((bookmarkResource) => {
    const urlSearchParams = new URLSearchParams(
      dataTransformer(bookmarkResource.data ?? ""),
    );

    const bookmark = <BookmarkEntry>{
      resource: bookmarkResource,
      absoluteTimeRange: isAbsoluteTimeRangeBookmark(urlSearchParams),
      filtersOnly: isFilterOnlyBookmark(urlSearchParams),
      url: "?" + urlSearchParams.toString(),
    };

    return bookmark;
  });
}

export function categorizeBookmarks(bookmarkEntries: BookmarkEntry[]) {
  const bookmarks: Bookmarks = {
    home: undefined,
    personal: [],
    shared: [],
  };

  bookmarkEntries.forEach((bookmark) => {
    if (isHomeBookmark(bookmark)) {
      bookmarks.home = bookmark;
    } else if (bookmark.shared) {
      bookmarks.shared.push(bookmark);
    } else {
      bookmarks.personal.push(bookmark);
    }
  });

  return bookmarks;
}

export function searchBookmarks(
  bookmarks: Bookmarks | undefined,
  searchText: string,
): Bookmarks | undefined {
  if (!searchText || !bookmarks) return bookmarks;
  searchText = searchText.toLowerCase();
  const matchName = (bookmark: BookmarkEntry | undefined) =>
    bookmark?.resource.displayName &&
    bookmark.resource.displayName.toLowerCase().includes(searchText);
  return {
    home: matchName(bookmarks.home) ? bookmarks.home : undefined,
    personal: bookmarks?.personal.filter(matchName) ?? [],
    shared: bookmarks?.shared.filter(matchName) ?? [],
  };
}

export function isBookmarkActive(entry: BookmarkEntry, curUrl: URL) {
  const bookmarkUrl = new URL(entry.url);

  if (entry.filtersOnly) {
    return bookmarkUrl.searchParams.entries().every(([key, value]) => {
      const curValue = curUrl.searchParams.get(key);
      return curValue === value;
    });
  }

  return bookmarkUrl.search === curUrl.search;
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

function isFilterOnlyBookmark(bookmarkUrlParams: URLSearchParams): boolean {
  const urlParams = Array.from(bookmarkUrlParams.keys());
  return urlParams.every((param) => filterOnlyParams.has(param));
}
