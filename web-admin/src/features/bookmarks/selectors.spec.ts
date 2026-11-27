import { getUrlParamsFromHomeBookmark } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
import { compressUrlParams } from "@rilldata/web-common/features/dashboards/url-state/compression.ts";
import { describe, expect, it, vi } from "vitest";

vi.mock("$app/stores", () => ({ page: {} }));

describe("getUrlParamsFromHomeBookmark", () => {
  const paramsOf = (urlSearch: string) =>
    getUrlParamsFromHomeBookmark({ urlSearch }).toString();

  it("opens a bookmark saved without a filter unfiltered, whatever the explore's default filter", () => {
    // Bookmarks saved today hold an empty filter param. Older ones hold none.
    // The empty param keeps the explore's default filter from filling it in.
    expect(paramsOf("tr=P7D&f=")).toEqual("tr=P7D&f=");
    expect(paramsOf("tr=P7D")).toEqual("tr=P7D&f=");
  });

  it("keeps the filter the bookmark was saved with", () => {
    expect(paramsOf("tr=P7D&f=domain IN ('fabrikam.com')")).toEqual(
      new URLSearchParams("tr=P7D&f=domain IN ('fabrikam.com')").toString(),
    );
  });

  it("expands a compressed bookmark before checking for a filter", () => {
    const compressed = new URLSearchParams({
      gzipped_state: compressUrlParams("tr=P7D&f=domain IN ('fabrikam.com')"),
    });
    expect(paramsOf(compressed.toString())).toEqual(
      new URLSearchParams("tr=P7D&f=domain IN ('fabrikam.com')").toString(),
    );
  });
});
