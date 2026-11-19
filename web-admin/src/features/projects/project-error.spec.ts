import { describe, expect, it } from "vitest";
import { resolveProjectError } from "./project-error";

// The AxiosError the admin client throws when the request never got a response
// (offline, DNS failure, timeout). There is no `response` to read a status or message from.
const networkError = {
  name: "AxiosError",
  message: "Network Error",
  code: "ERR_NETWORK",
  response: undefined,
};

function httpError(status: number, message: string) {
  return {
    name: "AxiosError",
    message: `Request failed with status code ${status}`,
    code: status >= 500 ? "ERR_BAD_RESPONSE" : "ERR_BAD_REQUEST",
    status,
    response: { status, data: { code: 2, message } },
  };
}

describe("resolveProjectError", () => {
  it("returns nothing while the query has no error", () => {
    expect(resolveProjectError(null, false)).toBeUndefined();
    expect(resolveProjectError(null, true)).toBeUndefined();
  });

  describe("without a cached project (the initial load failed)", () => {
    it("shows the server's status and message", () => {
      expect(
        resolveProjectError(httpError(404, "project not found"), false),
      ).toEqual({ statusCode: 404, message: "project not found" });
    });

    it("shows a network error without a status instead of crashing", () => {
      expect(resolveProjectError(networkError, false)).toEqual({
        statusCode: undefined,
        message: "Network Error",
      });
    });
  });

  describe("with a cached project (a background refetch failed)", () => {
    it("keeps the cached project on a network error", () => {
      expect(resolveProjectError(networkError, true)).toBeUndefined();
    });

    it("keeps the cached project on a server error", () => {
      for (const status of [500, 502, 503, 504]) {
        expect(
          resolveProjectError(httpError(status, "internal error"), true),
        ).toBeUndefined();
      }
    });

    it("keeps the cached project when rate limited", () => {
      expect(
        resolveProjectError(httpError(429, "too many requests"), true),
      ).toBeUndefined();
    });

    it("replaces the cached project once the user can no longer reach it", () => {
      expect(
        resolveProjectError(httpError(401, "auth token is expired"), true),
      ).toEqual({ statusCode: 401, message: "auth token is expired" });
      expect(
        resolveProjectError(httpError(403, "permission denied"), true),
      ).toEqual({ statusCode: 403, message: "permission denied" });
      expect(
        resolveProjectError(httpError(404, "project not found"), true),
      ).toEqual({ statusCode: 404, message: "project not found" });
    });
  });
});
