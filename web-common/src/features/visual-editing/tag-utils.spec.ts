import { describe, expect, it } from "vitest";
import { parseDocument } from "yaml";
import {
  getResourceTagSuggestions,
  readRootYamlTags,
  setRootYamlTags,
} from "./tag-utils";

describe("visual editing tag utils", () => {
  it("reads top-level YAML tags from scalar nodes", () => {
    const document = parseDocument(`
type: metrics_view
tags:
  - finance
  - "growth"
  - ""
  - finance
`);

    expect(readRootYamlTags(document)).toEqual(["finance", "growth"]);
  });

  it("writes and removes top-level YAML tags", () => {
    const document = parseDocument("type: explore\n");

    setRootYamlTags(document, [" analytics ", "growth", "analytics"]);
    expect(readRootYamlTags(document)).toEqual(["analytics", "growth"]);

    setRootYamlTags(document, []);
    expect(document.has("tags")).toBe(false);
  });

  it("builds suggestions from resource metadata and extra tags", () => {
    expect(
      getResourceTagSuggestions(
        [
          { meta: { tags: ["finance", "growth"] } },
          { meta: { tags: ["growth", "sales"] } },
        ] as never,
        ["executive"],
      ),
    ).toEqual(["executive", "finance", "growth", "sales"]);
  });
});
