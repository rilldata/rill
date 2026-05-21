import { describe, it, expect, vi } from "vitest";
import { rewriteCitationUrls } from "@rilldata/web-common/features/chat/core/messages/text/rewrite-citation-urls.ts";

const MAPPED = "http://localhost:3000/explore/AdBids_explore?view=explore";

describe("rewriteCitationUrls", () => {
  it("returns content unchanged when there are no citation URLs", async () => {
    const mapper = vi.fn(async (u: URL) => u.href);
    const content = "no links here, just text and a code `snippet`.";
    const result = await rewriteCitationUrls(content, mapper);
    expect(result).toBe(content);
    expect(mapper).not.toHaveBeenCalled();
  });

  it("rewrites a single absolute citation URL inside a markdown link", async () => {
    const mapper = vi.fn(async () => MAPPED);
    const content =
      "See [chart](http://localhost:3000/-/ai/sess/message/abc/-/open) for details.";
    const result = await rewriteCitationUrls(content, mapper);
    expect(result).toBe(`See [chart](${MAPPED}) for details.`);
    expect(mapper).toHaveBeenCalledTimes(1);
  });

  it("rewrites a relative citation URL and preserves relativity", async () => {
    const mapper = vi.fn(
      async () => "http://localhost/explore/foo?view=explore",
    );
    const content = "Click [here](/-/ai/sess/message/abc/-/open).";
    const result = await rewriteCitationUrls(content, mapper);
    expect(result).toBe("Click [here](/explore/foo?view=explore).");
  });

  it("rewrites legacy /-/open-query citation URLs", async () => {
    const mapper = vi.fn(async () => MAPPED);
    const content =
      "[legacy](http://localhost:3000/-/open-query?query=%7B%7D)";
    const result = await rewriteCitationUrls(content, mapper);
    expect(result).toBe(`[legacy](${MAPPED})`);
  });

  it("rewrites multiple citation URLs in parallel", async () => {
    const mapper = vi.fn(async (url: URL) => `${MAPPED}#${url.pathname}`);
    const content =
      "First [a](http://localhost:3000/-/ai/s/message/one/-/open) and " +
      "second [b](http://localhost:3000/-/ai/s/message/two/-/open).";
    const result = await rewriteCitationUrls(content, mapper);
    expect(mapper).toHaveBeenCalledTimes(2);
    expect(result).toBe(
      `First [a](${MAPPED}#/-/ai/s/message/one/-/open) and ` +
        `second [b](${MAPPED}#/-/ai/s/message/two/-/open).`,
    );
  });

  it("leaves non-citation URLs untouched", async () => {
    const mapper = vi.fn(async () => MAPPED);
    const content =
      "Visit [docs](https://docs.rilldata.com/intro) or [home](/dashboards).";
    const result = await rewriteCitationUrls(content, mapper);
    expect(result).toBe(content);
    expect(mapper).not.toHaveBeenCalled();
  });

  it("handles a mix of citation and non-citation URLs", async () => {
    const mapper = vi.fn(async () => MAPPED);
    const content =
      "[docs](https://docs.rilldata.com) then " +
      "[chart](http://localhost:3000/-/ai/sess/message/abc/-/open).";
    const result = await rewriteCitationUrls(content, mapper);
    expect(mapper).toHaveBeenCalledTimes(1);
    expect(result).toBe(`[docs](https://docs.rilldata.com) then [chart](${MAPPED}).`);
  });
});
