// @vitest-environment jsdom
import { afterEach, describe, expect, it } from "vitest";
import {
  clearPublicURLSession,
  getPublicURLSession,
  setPublicURLSession,
} from "./password-session";

describe("password-session", () => {
  afterEach(() => {
    sessionStorage.clear();
  });

  it("returns a stored session until it expires", () => {
    const expiresOn = new Date(Date.now() + 60_000).toISOString();
    setPublicURLSession("rill_mgc_a", { token: "rill_mgs_a", expiresOn });
    expect(getPublicURLSession("rill_mgc_a")).toEqual({
      token: "rill_mgs_a",
      expiresOn,
    });
    expect(getPublicURLSession("rill_mgc_other")).toBeUndefined();
  });

  it("discards expired sessions", () => {
    setPublicURLSession("rill_mgc_a", {
      token: "rill_mgs_a",
      expiresOn: new Date(Date.now() - 1_000).toISOString(),
    });
    expect(getPublicURLSession("rill_mgc_a")).toBeUndefined();
    expect(sessionStorage.length).toBe(0);
  });

  it("discards unreadable entries", () => {
    sessionStorage.setItem("rill:public-url-session:rill_mgc_a", "{not json");
    expect(getPublicURLSession("rill_mgc_a")).toBeUndefined();
    expect(sessionStorage.length).toBe(0);
  });

  it("clears a session", () => {
    setPublicURLSession("rill_mgc_a", {
      token: "rill_mgs_a",
      expiresOn: new Date(Date.now() + 60_000).toISOString(),
    });
    clearPublicURLSession("rill_mgc_a");
    expect(getPublicURLSession("rill_mgc_a")).toBeUndefined();
  });
});
