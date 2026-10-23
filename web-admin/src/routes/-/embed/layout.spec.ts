import { isHttpError, isRedirect } from "@sveltejs/kit";
import { describe, expect, it, vi } from "vitest";
import { load } from "./+layout";

// Each load call is treated as the first one in the iframe, so it always handles the entry redirect.
vi.mock("@rilldata/web-common/features/embeds/embed-store.ts", () => ({
  EmbedStore: { getInstance: () => null, init: vi.fn() },
}));
// The real module needs a browser environment; only `removeEmbedParams` is used by the layout.
vi.mock("@rilldata/web-admin/features/embeds/init-embed-public-api.ts", () => ({
  removeEmbedParams: (searchParams: URLSearchParams) => {
    const cleaned = new URLSearchParams(searchParams);
    ["instance_id", "runtime_host", "access_token", "resource", "type"].forEach(
      (p) => cleaned.delete(p),
    );
    return cleaned.toString();
  },
}));

function encodeOpen(payload: unknown): string {
  return btoa(JSON.stringify(payload))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

async function callLoad(search: Record<string, string>): Promise<unknown> {
  const url = new URL("https://ui.rilldata.com/-/embed");
  url.search = new URLSearchParams({
    instance_id: "inst",
    runtime_host: "https://runtime.rilldata.com",
    access_token: "tok",
    ...search,
  }).toString();
  try {
    await load({ url } as never);
    return undefined;
  } catch (e) {
    return e;
  }
}

async function expectRedirect(
  search: Record<string, string>,
  location: string,
) {
  const result = await callLoad(search);
  expect(isRedirect(result)).toBe(true);
  if (!isRedirect(result)) return;
  expect(result.status).toBe(307);
  expect(result.location).toBe(location);
}

describe("embed/+layout load", () => {
  it("redirects to the report open route", async () => {
    await expectRedirect(
      {
        open: encodeOpen({
          kind: "report",
          name: "weekly_revenue",
          execution_time: "2026-10-20T00:00:00Z",
        }),
      },
      "/-/embed/reports/weekly_revenue/open?execution_time=2026-10-20T00%3A00%3A00Z",
    );
  });

  it("drops other params when opening a report", async () => {
    await expectRedirect(
      {
        open: encodeOpen({ kind: "report", name: "r" }),
        navigation: "true",
        tr: "P7D",
      },
      "/-/embed/reports/r/open",
    );
  });

  it("errors on an invalid open param", async () => {
    const result = await callLoad({ open: "not-valid" });
    expect(isHttpError(result)).toBe(true);
    if (!isHttpError(result)) return;
    expect(result.status).toBe(400);
  });

  it("still redirects to the dashboard without an open param", async () => {
    await expectRedirect(
      { resource: "e", type: "rill.runtime.v1.Explore", tr: "P7D" },
      "/-/embed/explore/e?tr=P7D",
    );
  });
});
