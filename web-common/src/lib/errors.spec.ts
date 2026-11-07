import { describe, expect, it } from "vitest";
import { isNetworkError } from "./errors";

describe("isNetworkError", () => {
  it("matches Axios network errors", () => {
    expect(isNetworkError(new Error("Network Error"))).toBe(true);
    expect(isNetworkError({ code: "ERR_NETWORK" })).toBe(true);
  });

  it("matches browser fetch transport failures", () => {
    expect(isNetworkError(new TypeError("Failed to fetch"))).toBe(true);
    expect(isNetworkError(new TypeError("Load failed"))).toBe(true);
    expect(isNetworkError(new Error("fetch failed"))).toBe(true);
  });

  it("matches Connect unavailable errors that wrap fetch failures", () => {
    expect(isNetworkError({ code: 14, rawMessage: "fetch failed" })).toBe(true);
  });

  it("does not treat HTTP responses as network errors", () => {
    expect(
      isNetworkError({
        response: {
          status: 503,
          data: { message: "Network Error" },
        },
      }),
    ).toBe(false);
    expect(
      isNetworkError({ status: 503, message: "Service unavailable" }),
    ).toBe(false);
    expect(isNetworkError({ code: 14, rawMessage: "unavailable" })).toBe(false);
  });
});
