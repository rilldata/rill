// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getPublicURLSession, setPublicURLSession } from "./password-session";
import { resolvePublicURLBearer } from "./public-url-bearer";

const mocks = vi.hoisted(() => ({ fetch: vi.fn() }));
vi.mock("@rilldata/web-admin/features/projects/selectors", () => ({
  fetchMagicAuthToken: mocks.fetch,
}));

const session = {
  token: "rill_mgs_a",
  expiresOn: new Date(Date.now() + 60_000).toISOString(),
};

describe("resolvePublicURLBearer", () => {
  beforeEach(() => {
    setPublicURLSession("rill_mgc_a", session);
  });
  afterEach(() => {
    sessionStorage.clear();
    vi.clearAllMocks();
  });

  it("uses the stored session as the bearer", async () => {
    mocks.fetch.mockResolvedValue({ token: { id: "a" } });
    expect(await resolvePublicURLBearer("rill_mgc_a")).toEqual({
      bearer: "rill_mgs_a",
      locked: false,
      tokenMetadata: { id: "a" },
    });
  });

  it("discards a session that is rejected and locks the URL again", async () => {
    mocks.fetch.mockImplementation((token: string) =>
      token === "rill_mgs_a"
        ? Promise.reject({ response: { status: 401 } })
        : Promise.resolve({ token: { id: "a" }, locked: true }),
    );
    expect(await resolvePublicURLBearer("rill_mgc_a")).toMatchObject({
      bearer: "rill_mgc_a",
      locked: true,
    });
    expect(getPublicURLSession("rill_mgc_a")).toBeUndefined();
  });

  it.each([429, 503, undefined])(
    "keeps the session when the request fails with status %s",
    async (status) => {
      const failure = status ? { response: { status } } : new Error("offline");
      mocks.fetch.mockRejectedValue(failure);
      await expect(resolvePublicURLBearer("rill_mgc_a")).rejects.toBe(failure);
      expect(getPublicURLSession("rill_mgc_a")).toEqual(session);
    },
  );
});
