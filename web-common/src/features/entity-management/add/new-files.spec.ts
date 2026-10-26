import { describe, expect, it } from "vitest";
import { ResourceKind } from "../resource-selectors.ts";
import { getPathForNewResourceFile } from "./new-files.ts";

describe("getPathForNewResourceFile", () => {
  it("uses the current directory when creating a component", () => {
    expect(
      getPathForNewResourceFile(ResourceKind.Component, undefined, "Something"),
    ).toMatch(/^Something\/component/);
  });
});
