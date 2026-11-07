import { describe, expect, it } from "vitest";
import { usePageOriginForLocalhost } from "./url-utils";

describe("usePageOriginForLocalhost", () => {
  it("rewrites a localhost login URL to the page host", () => {
    expect(
      usePageOriginForLocalhost(
        "http://localhost:9009/auth",
        "http://rill.example:9009/files",
      ),
    ).toBe("http://rill.example:9009/auth");
  });

  it("leaves a non-local login URL unchanged", () => {
    const cloud = "https://admin.rilldata.com/auth";
    expect(
      usePageOriginForLocalhost(cloud, "http://rill.example:9009/files"),
    ).toBe(cloud);
  });

  it("leaves an already matching localhost URL unchanged", () => {
    const same = "http://localhost:9009/auth";
    expect(usePageOriginForLocalhost(same, "http://localhost:9009/")).toBe(
      same,
    );
  });
});
