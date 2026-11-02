import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { isHttpError, isRedirect } from "@sveltejs/kit";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { load } from "./+layout";

// The real module sets up the embed's public API on import, which requires a browser.
// Only `removeEmbedParams` is used by the layout.
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

describe("embed/+layout load on the first load", () => {
  // Each load call is treated as the first one in the iframe, so it always handles the entry redirect.
  beforeEach(() => {
    vi.spyOn(EmbedStore, "getInstance").mockReturnValue(null);
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

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

async function loadWithParams(params: Record<string, string>) {
  const url = new URL("http://localhost/-/embed/explore/e1");
  for (const [k, v] of Object.entries(params)) {
    url.searchParams.set(k, v);
  }
  EmbedStore.init(url);
  return load({ url } as never);
}

const requiredParams = {
  instance_id: "i1",
  runtime_host: "http://runtime.example.com",
  access_token: "token",
};

describe("embed/+layout load after the first load", () => {
  it("enables admin actions with an admin token and no user email", async () => {
    const data = await loadWithParams({
      ...requiredParams,
      admin_token: "admin-token",
    });
    expect(data.embedAdminActions).toBe(true);
  });

  it("disables admin actions without an admin token", async () => {
    const data = await loadWithParams({
      ...requiredParams,
      user_email: "embed@example.com",
    });
    expect(data.embedAdminActions).toBe(false);
  });
});
