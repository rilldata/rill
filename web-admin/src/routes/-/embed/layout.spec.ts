import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { describe, expect, it, vi } from "vitest";
import { load } from "./+layout";

// The module sets up the embed's public API on import, which requires a browser.
vi.mock("@rilldata/web-admin/features/embeds/init-embed-public-api.ts", () => ({
  removeEmbedParams: (params: URLSearchParams) => params,
}));

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

describe("embed/+layout load", () => {
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
