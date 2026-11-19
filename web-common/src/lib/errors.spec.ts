import { Code, ConnectError } from "@connectrpc/connect";
import { CancelledError } from "@tanstack/query-core";
import { describe, expect, it } from "vitest";
import { isCancelledRequestError, isNetworkError } from "./errors";

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

describe("isCancelledRequestError", () => {
  it("matches aborted fetches and the Connect errors that wrap them", () => {
    expect(
      isCancelledRequestError(
        new DOMException("signal is aborted without reason", "AbortError"),
      ),
    ).toBe(true);
    // In browsers DOMException extends Error, which is what ConnectError.from keys on; jsdom's does not.
    const abort = new Error("The user aborted a request.");
    abort.name = "AbortError";
    expect(isCancelledRequestError(ConnectError.from(abort))).toBe(true);
    expect(
      isCancelledRequestError(
        new ConnectError("context canceled", Code.Canceled),
      ),
    ).toBe(true);
  });

  it("matches axios cancellations", () => {
    expect(
      isCancelledRequestError({ code: "ERR_CANCELED", message: "canceled" }),
    ).toBe(true);
    expect(
      isCancelledRequestError({
        code: "ECONNABORTED",
        message: "Request aborted",
      }),
    ).toBe(true);
  });

  it("matches TanStack Query cancellations", () => {
    expect(isCancelledRequestError(new CancelledError())).toBe(true);
  });

  it("does not match timeouts or other failures", () => {
    expect(
      isCancelledRequestError(
        new ConnectError("context deadline exceeded", Code.DeadlineExceeded),
      ),
    ).toBe(false);
    expect(
      isCancelledRequestError({
        code: "ECONNABORTED",
        message: "timeout of 30000ms exceeded",
      }),
    ).toBe(false);
    // A DOMException with code 1 is not a Connect Canceled error
    expect(
      isCancelledRequestError(
        new DOMException("index out of range", "IndexSizeError"),
      ),
    ).toBe(false);
    // Server failures and plain errors that merely mention cancellation are not client cancellations
    expect(
      isCancelledRequestError(
        new ConnectError(
          "reconciles for resources have hung for more than 10m0s after cancelation",
          Code.Unknown,
        ),
      ),
    ).toBe(false);
    expect(
      isCancelledRequestError({
        response: { status: 500, data: { message: "query was cancelled" } },
      }),
    ).toBe(false);
    expect(
      isCancelledRequestError(new Error("the user cancelled the export")),
    ).toBe(false);
    expect(isCancelledRequestError("AbortError")).toBe(false);
    expect(isCancelledRequestError(undefined)).toBe(false);
  });
});
