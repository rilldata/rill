import {
  removeFieldsFromFilterConfig,
  setPinnedAndRequiredFilters,
} from "@rilldata/web-common/features/dashboards/filters/utils.ts";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { V1Operation } from "@rilldata/web-common/runtime-client";
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

describe("removeFieldsFromFilterConfig", () => {
  // What the runtime parses from the `defaults.filters` of the documents below.
  const defaultFilter = createAndExpression([
    createInExpression("country", ["US"]),
    createSubQueryExpression(
      "publisher",
      ["impressions"],
      createBinaryExpression("impressions", V1Operation.OPERATION_GT, 10),
    ),
  ]);
  const explore = `type: explore
filters:
  pinned: [domain, publisher]
  required: [publisher]
defaults:
  time_range: P7D
  filters: country IN ('US') AND publisher IN (SELECT publisher FROM ads HAVING impressions > 10)
`;

  it("takes the fields out of the lists and leaves a default filter that does not use them", () => {
    const yaml = parseDocument(explore);

    removeFieldsFromFilterConfig(yaml, new Set(["domain"]), defaultFilter);
    expect(yaml.toJSON()).toEqual({
      ...parseDocument(explore).toJSON(),
      filters: { pinned: ["publisher"], required: ["publisher"] },
    });
  });

  it("removes a default filter that uses one of the fields, also as the measure of a measure filter", () => {
    for (const field of ["publisher", "impressions"]) {
      const yaml = parseDocument(
        `type: metrics_view\nexplore:\n${explore.replace(/^/gm, "  ")}`,
      );

      removeFieldsFromFilterConfig(yaml, new Set([field]), defaultFilter, [
        "explore",
      ]);
      expect(yaml.toJSON().explore.defaults).toEqual({ time_range: "P7D" });
    }
  });

  it("leaves the document as it is when it does not use the fields", () => {
    const yaml = parseDocument(explore);

    removeFieldsFromFilterConfig(yaml, new Set(["city"]), defaultFilter);
    expect(yaml.toString()).toEqual(parseDocument(explore).toString());
  });
});
