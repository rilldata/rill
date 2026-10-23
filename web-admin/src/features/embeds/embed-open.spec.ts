import { describe, expect, it } from "vitest";
import { decodeEmbedOpen, getEmbedOpenPath } from "./embed-open";

// Mirrors embedopen.Encode in the runtime: unpadded base64url of a JSON object.
function encode(payload: unknown): string {
  const bytes = new TextEncoder().encode(JSON.stringify(payload));
  return btoa(String.fromCharCode(...bytes))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

describe("decodeEmbedOpen", () => {
  it("decodes a report payload with an execution time", () => {
    expect(
      decodeEmbedOpen(
        encode({
          kind: "report",
          name: "weekly_revenue",
          execution_time: "2026-10-20T00:00:00Z",
        }),
      ),
    ).toEqual({
      kind: "report",
      name: "weekly_revenue",
      executionTime: "2026-10-20T00:00:00Z",
    });
  });

  it("decodes a payload without an execution time", () => {
    expect(decodeEmbedOpen(encode({ kind: "report", name: "r" }))).toEqual({
      kind: "report",
      name: "r",
    });
  });

  it("decodes base64url characters and non-ASCII names", () => {
    // "ü" and "?>" produce "-" and "_" in base64url.
    const blob = encode({ kind: "report", name: "ü?>?>" });
    expect(blob).toMatch(/[-_]/);
    expect(decodeEmbedOpen(blob)).toEqual({ kind: "report", name: "ü?>?>" });
  });

  it("accepts padded input", () => {
    const blob = encode({ kind: "report", name: "r" });
    const padded = blob.padEnd(Math.ceil(blob.length / 4) * 4, "=");
    expect(decodeEmbedOpen(padded)).toEqual({ kind: "report", name: "r" });
  });

  it("ignores unknown fields", () => {
    expect(
      decodeEmbedOpen(encode({ kind: "report", name: "r", extra: 1 })),
    ).toEqual({ kind: "report", name: "r" });
  });

  it.each([
    ["not base64", "!!!"],
    ["not JSON", encode("x").slice(0, 2)],
    ["a JSON string", encode("report")],
    ["null", encode(null)],
    ["unknown kind", encode({ kind: "canvas", name: "c" })],
    ["missing name", encode({ kind: "report" })],
    ["empty name", encode({ kind: "report", name: "" })],
    ["non-string name", encode({ kind: "report", name: 1 })],
    [
      "invalid execution time",
      encode({ kind: "report", name: "r", execution_time: "yesterday" }),
    ],
    [
      "non-string execution time",
      encode({ kind: "report", name: "r", execution_time: 1 }),
    ],
  ])("rejects %s", (_, blob) => {
    expect(decodeEmbedOpen(blob)).toBeNull();
  });
});

describe("getEmbedOpenPath", () => {
  it("builds the report open route with the execution time", () => {
    expect(
      getEmbedOpenPath({
        kind: "report",
        name: "weekly_revenue",
        executionTime: "2026-10-20T00:00:00Z",
      }),
    ).toBe(
      "/-/embed/reports/weekly_revenue/open?execution_time=2026-10-20T00%3A00%3A00Z",
    );
  });

  it("builds the open route without an execution time", () => {
    expect(getEmbedOpenPath({ kind: "report", name: "a b" })).toBe(
      "/-/embed/reports/a%20b/open",
    );
  });
});
