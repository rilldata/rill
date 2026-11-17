import { setPinnedAndRequiredFilters } from "@rilldata/web-common/features/dashboards/filters/utils.ts";
import { describe, expect, it } from "vitest";
import { parseDocument } from "yaml";

describe("setPinnedAndRequiredFilters", () => {
  it("writes the lists, removes the empty one and leaves the other filter keys", () => {
    const yaml = parseDocument(
      "type: canvas\nfilters:\n  enable: true\n  required: [publisher]\n",
    );

    setPinnedAndRequiredFilters(yaml, ["domain"], []);

    expect(yaml.toJSON()).toEqual({
      type: "canvas",
      filters: { enable: true, pinned: ["domain"] },
    });
  });

  it("writes under the key path of an explore defined in a metrics view", () => {
    const yaml = parseDocument(
      "type: metrics_view\nexplore:\n  display_name: Ads\n",
    );

    setPinnedAndRequiredFilters(yaml, [], ["publisher"], ["explore"]);
    expect(yaml.toJSON()).toEqual({
      type: "metrics_view",
      explore: { display_name: "Ads", filters: { required: ["publisher"] } },
    });

    // The block goes away once nothing is left in it.
    setPinnedAndRequiredFilters(yaml, [], [], ["explore"]);
    expect(yaml.toJSON()).toEqual({
      type: "metrics_view",
      explore: { display_name: "Ads" },
    });
  });
});
