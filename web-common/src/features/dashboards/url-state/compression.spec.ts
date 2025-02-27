import {
  RandomDomains,
  RandomPublishers,
} from "@rilldata/web-common/features/dashboards/stores/test-data/random";
import {
  compressUrlParams,
  decompressUrlParams,
} from "@rilldata/web-common/features/dashboards/url-state/compression";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { describe, expect, it } from "vitest";

describe("URL Compression", () => {
  it("Small URL", async () => {
    const url = new URL(`http://localhost/explore/AdBids`);
    url.searchParams.set("f", "publisher in ('Google', 'Yahoo')");

    const compressedUrl = new URL(`http://localhost/explore/AdBids`);
    compressedUrl.search = await compressUrlParams(url);

    // compressed url is the same as the original since the url is small enough
    expect(url.toString()).toEqual(compressedUrl.toString());
  });

  it("Large URL", async () => {
    const url = new URL(`http://localhost/explore/AdBids`);
    url.searchParams.set(
      "f",
      `publisher in (${RandomPublishers.map((p) => `"${p}"`).join(",")}) and` +
        `domain in (${RandomDomains.map((d) => `"${d}"`).join(",")})`,
    );

    const compressedUrl = new URL(`http://localhost/explore/AdBids`);
    compressedUrl.search = await compressUrlParams(url);

    // compressed url is not the same as the original
    expect(url.toString()).not.toEqual(compressedUrl.toString());
    // compressed url has p_gzip
    expect(compressedUrl.searchParams.has(ExploreStateURLParams.GzippedParams));

    const decompressedUrl = new URL(`http://localhost/explore/AdBids`);
    decompressedUrl.search = (
      await decompressUrlParams(compressedUrl.searchParams)
    ).toString();
    // after decompressing url matches the original
    expect(url.toString()).toEqual(decompressedUrl.toString());
  });
});
