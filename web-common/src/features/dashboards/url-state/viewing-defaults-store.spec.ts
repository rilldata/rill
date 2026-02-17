import { describe, it, expect } from "vitest";
import { cleanUrlParams } from "web-common/src/features/dashboards/url-state/clean-url-params";
import { isViewingDefaults } from "web-common/src/features/dashboards/url-state/viewing-defaults-store";
import { getExploreStateFromYAMLConfig } from "web-common/src/features/dashboards/stores/get-explore-state-from-yaml-config";
import { convertPartialExploreStateToUrlParams } from "web-common/src/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { getRillDefaultExploreUrlParams } from "web-common/src/features/dashboards/url-state/get-rill-default-explore-url-params";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1TimeRangeSummary,
} from "web-common/src/runtime-client";
import {
  MetricsViewSpecDimensionType,
  V1ExploreComparisonMode,
  V1Operation,
  V1TimeGrain,
} from "web-common/src/runtime-client";
import type { TimeControlState } from "web-common/src/features/dashboards/time-controls/time-control-store";

// ─── Shared fixtures ───────────────────────────────────────────────

const MEASURES = [
  "total_cost",
  "total_revenue",
  "net_revenue",
  "gross_margin_percent",
  "unique_customers",
];
const DIMENSIONS = [
  "customer",
  "plan_name",
  "location",
  "component",
  "app_name",
  "sku_description",
  "pipeline",
  "environment",
];
const ALL_DIMENSIONS = ["__time", ...DIMENSIONS];

const TIME_RANGE_SUMMARY: V1TimeRangeSummary = {
  min: "2024-01-01T00:00:00Z",
  max: "2025-01-01T00:00:00Z",
};

const METRICS_VIEW_SPEC: V1MetricsViewSpec = {
  measures: MEASURES.map((name) => ({ name })),
  dimensions: [
    { name: "__time", type: MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME },
    ...DIMENSIONS.map((name) => ({
      name,
      type: MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    })),
  ],
  timeDimension: "__time",
  smallestTimeGrain: V1TimeGrain.TIME_GRAIN_DAY,
};

/** Base explore spec — all measures/dimensions, no defaults */
function makeExploreSpec(
  defaultPreset: V1ExploreSpec["defaultPreset"] = {},
): V1ExploreSpec {
  return {
    measures: MEASURES,
    dimensions: ALL_DIMENSIONS,
    defaultPreset: {
      measures: MEASURES,
      dimensions: ALL_DIMENSIONS,
      ...defaultPreset,
    },
  };
}

// ─── Helpers ────────────────────────────────────────────────────────

/**
 * Converts a YAML-like explore spec into YAML default URL params,
 * the same way state-managers.ts does it (using the YAML explore state
 * as the TimeControlState to avoid isoDurationToFullTimeRange issues).
 */
function getYamlDefaultUrlParams(exploreSpec: V1ExploreSpec) {
  const yamlState = getExploreStateFromYAMLConfig(
    exploreSpec,
    TIME_RANGE_SUMMARY,
    METRICS_VIEW_SPEC.smallestTimeGrain,
  );
  const timeControlState: Partial<TimeControlState> = {
    selectedTimeRange: yamlState.selectedTimeRange,
    selectedComparisonTimeRange: yamlState.selectedComparisonTimeRange,
  };
  return convertPartialExploreStateToUrlParams(
    exploreSpec,
    METRICS_VIEW_SPEC,
    yamlState,
    timeControlState as TimeControlState,
  );
}

function getRillDefaultUrlParams(exploreSpec: V1ExploreSpec) {
  return getRillDefaultExploreUrlParams(
    METRICS_VIEW_SPEC,
    exploreSpec,
    TIME_RANGE_SUMMARY,
  );
}

/**
 * Simulates what DashboardStateSync does: given the full explore state
 * as URL params, clean them against rill defaults to produce the browser URL.
 */
function simulateBrowserUrl(
  fullStateParams: URLSearchParams,
  rillDefaults: URLSearchParams,
): URLSearchParams {
  return cleanUrlParams(fullStateParams, rillDefaults);
}

/**
 * End-to-end helper: given an explore spec, compute everything and test
 * whether a specific browser URL is detected as "viewing defaults".
 */
function checkViewingDefaults(
  exploreSpec: V1ExploreSpec,
  browserUrl: string,
): boolean {
  const yamlDefaults = getYamlDefaultUrlParams(exploreSpec);
  const rillDefaults = getRillDefaultUrlParams(exploreSpec);
  const raw = new URLSearchParams(browserUrl);
  const cleaned = cleanUrlParams(raw, yamlDefaults);
  return isViewingDefaults(cleaned, yamlDefaults, rillDefaults, raw);
}

// ═══════════════════════════════════════════════════════════════════
// Tests
// ═══════════════════════════════════════════════════════════════════

describe("YAML → URL params pipeline", () => {
  it("generates correct URL params for comparison_mode: rill-PQ", () => {
    const spec = makeExploreSpec({
      timeRange: "14D as of latest/D+1D",
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME,
      compareTimeRange: "rill-PQ",
    });
    const yamlState = getExploreStateFromYAMLConfig(
      spec,
      TIME_RANGE_SUMMARY,
      METRICS_VIEW_SPEC.smallestTimeGrain,
    );

    // Verify the explore state has the right comparison fields
    expect(yamlState.showTimeComparison).toBe(true);
    expect(yamlState.selectedComparisonTimeRange?.name).toBe("rill-PQ");
    expect(yamlState.selectedTimeRange?.name).toBe("14D as of latest/D+1D");

    // Convert to URL params
    const params = getYamlDefaultUrlParams(spec);
    expect(params.get("tr")).toBe("14D as of latest/D+1D");
    expect(params.get("compare_tr")).toBe("rill-PQ");
    expect(params.get("grain")).toBe("day");
  });

  it("generates correct URL params for comparison_mode: rill-PP (legacy 'time')", () => {
    // comparison_mode: time in YAML (no compareTimeRange) → falls back to rill-PP
    const spec = makeExploreSpec({
      timeRange: "P7D",
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME,
      // No compareTimeRange — simulates legacy "comparison_mode: time"
    });
    const yamlState = getExploreStateFromYAMLConfig(
      spec,
      TIME_RANGE_SUMMARY,
      METRICS_VIEW_SPEC.smallestTimeGrain,
    );

    expect(yamlState.showTimeComparison).toBe(true);
    // For P7D against 1-year data, the valid default comparison is rill-PW (previous week)
    expect(yamlState.selectedComparisonTimeRange?.name).toBe("rill-PW");
  });

  it("generates correct URL params for no comparison", () => {
    const spec = makeExploreSpec({
      timeRange: "P7D",
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE,
    });
    const yamlState = getExploreStateFromYAMLConfig(
      spec,
      TIME_RANGE_SUMMARY,
      METRICS_VIEW_SPEC.smallestTimeGrain,
    );

    expect(yamlState.showTimeComparison).toBeUndefined();
    expect(yamlState.selectedComparisonTimeRange).toBeUndefined();

    const params = getYamlDefaultUrlParams(spec);
    // NONE mode doesn't include selectedComparisonTimeRange in partial state,
    // so compare_tr isn't set in the URL params at all
    expect(params.has("compare_tr")).toBe(false);
  });

  it("generates filter param from defaultPreset.filter", () => {
    const spec = makeExploreSpec({
      timeRange: "P7D",
      filter: {
        expression: {
          cond: {
            op: V1Operation.OPERATION_AND,
            exprs: [
              {
                cond: {
                  op: V1Operation.OPERATION_IN,
                  exprs: [{ ident: "customer" }, { val: "Airtable" }],
                },
              },
              {
                cond: {
                  op: V1Operation.OPERATION_IN,
                  exprs: [{ ident: "location" }, { val: "us-east1" }],
                },
              },
            ],
          },
        },
      },
    });
    const params = getYamlDefaultUrlParams(spec);
    const filterParam = params.get("f") ?? "";

    // Filter should be non-empty and contain the dimension names
    expect(filterParam).not.toBe("");
    expect(filterParam).toContain("customer");
    expect(filterParam).toContain("Airtable");
    expect(filterParam).toContain("location");
    expect(filterParam).toContain("us-east1");
  });

  it("generates visible measures/dimensions correctly", () => {
    // When YAML measures match all spec measures → measures=*
    const specAll = makeExploreSpec({
      timeRange: "P7D",
    });
    const paramsAll = getYamlDefaultUrlParams(specAll);
    expect(paramsAll.get("measures")).toBe("*");
    expect(paramsAll.get("dims")).toBe("*");

    // When YAML has a subset of measures → specific list
    const specSubset = makeExploreSpec({
      timeRange: "P7D",
      measures: ["total_cost", "total_revenue"],
      dimensions: ["customer", "location"],
    });
    const paramsSubset = getYamlDefaultUrlParams(specSubset);
    expect(paramsSubset.get("measures")).toBe("total_cost,total_revenue");
    // __time is filtered out from dims since it's a time dimension
    expect(paramsSubset.get("dims")).toBe("customer,location");
  });
});

describe("isViewingDefaults — pure function", () => {
  describe("with no filters in YAML defaults", () => {
    // YAML: time_range + comparison_mode + all measures/dims (no filter)
    const spec = makeExploreSpec({
      timeRange: "14D as of latest/D+1D",
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME,
      compareTimeRange: "rill-PQ",
    });
    const yamlDefaults = getYamlDefaultUrlParams(spec);
    const rillDefaults = getRillDefaultUrlParams(spec);

    // The browser URL when truly viewing YAML defaults:
    // DashboardStateSync cleans the full state against rill defaults.
    const expectedBrowserUrl = simulateBrowserUrl(yamlDefaults, rillDefaults);

    it("returns true when browser URL matches YAML defaults", () => {
      const cleaned = cleanUrlParams(expectedBrowserUrl, yamlDefaults);
      expect(
        isViewingDefaults(
          cleaned,
          yamlDefaults,
          rillDefaults,
          expectedBrowserUrl,
        ),
      ).toBe(true);
    });

    it("returns true with the hardcoded URL the browser would show", () => {
      // Verify the expected browser URL has the params we expect
      expect(expectedBrowserUrl.get("tr")).toBe("14D as of latest/D+1D");
      expect(expectedBrowserUrl.get("compare_tr")).toBe("rill-PQ");
      expect(expectedBrowserUrl.has("grain")).toBe(true);
      // measures/dims should NOT be in the browser URL (cleaned because they match rill defaults)
      expect(expectedBrowserUrl.has("measures")).toBe(false);
      expect(expectedBrowserUrl.has("dims")).toBe(false);

      // Check viewing defaults with this URL
      expect(checkViewingDefaults(spec, expectedBrowserUrl.toString())).toBe(
        true,
      );
    });

    it("returns false when browser URL has a different time range", () => {
      const browserUrl = new URLSearchParams(expectedBrowserUrl);
      browserUrl.set("tr", "P7D");
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(false);
    });

    it("returns false when browser URL has extra params", () => {
      const browserUrl = new URLSearchParams(expectedBrowserUrl);
      browserUrl.set("expand_dim", "customer");
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(false);
    });
  });

  describe("with filters in YAML defaults", () => {
    const filterExpr = {
      cond: {
        op: V1Operation.OPERATION_AND,
        exprs: [
          {
            cond: {
              op: V1Operation.OPERATION_IN,
              exprs: [{ ident: "customer" }, { val: "Airtable" }],
            },
          },
          {
            cond: {
              op: V1Operation.OPERATION_IN,
              exprs: [{ ident: "location" }, { val: "us-east1" }],
            },
          },
        ],
      },
    };

    const spec = makeExploreSpec({
      timeRange: "14D as of latest/D+1D",
      comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME,
      compareTimeRange: "rill-PQ",
      filter: { expression: filterExpr },
    });
    const yamlDefaults = getYamlDefaultUrlParams(spec);
    const rillDefaults = getRillDefaultUrlParams(spec);

    // The filter should be in the YAML default params
    const yamlFilterParam = yamlDefaults.get("f") ?? "";

    it("YAML defaults include a non-empty filter param", () => {
      expect(yamlFilterParam).not.toBe("");
      expect(yamlFilterParam).toContain("customer");
    });

    it("returns true when browser URL has matching filter", () => {
      // Full viewing-defaults browser URL includes the filter
      const expectedBrowserUrl = simulateBrowserUrl(yamlDefaults, rillDefaults);
      expect(expectedBrowserUrl.has("f")).toBe(true);

      const cleaned = cleanUrlParams(expectedBrowserUrl, yamlDefaults);
      expect(
        isViewingDefaults(
          cleaned,
          yamlDefaults,
          rillDefaults,
          expectedBrowserUrl,
        ),
      ).toBe(true);
    });

    it("returns false when browser URL is missing the filter (the original bug)", () => {
      // This is the exact scenario from the bug report:
      // YAML has filters, browser URL has tr/compare_tr/grain but NO f param
      const browserUrl = new URLSearchParams(
        "tr=14D+as+of+latest%2FD%2B1D&compare_tr=rill-PQ&grain=day",
      );
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);

      // Forward check passes — all browser params match YAML defaults
      expect(cleaned.size).toBe(0);

      // But isViewingDefaults should catch the missing filter
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(false);
    });

    it("returns false when browser URL has a different filter", () => {
      const expectedBrowserUrl = simulateBrowserUrl(yamlDefaults, rillDefaults);
      const browserUrl = new URLSearchParams(expectedBrowserUrl);
      browserUrl.set("f", "location IN ('eu-west1')");
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(false);
    });
  });

  describe("edge cases", () => {
    it("returns false when yamlDefaults is undefined", () => {
      expect(
        isViewingDefaults(
          new URLSearchParams(),
          undefined,
          new URLSearchParams(),
          new URLSearchParams(),
        ),
      ).toBe(false);
    });

    it("returns false when yamlDefaults is empty", () => {
      expect(
        isViewingDefaults(
          new URLSearchParams(),
          new URLSearchParams(),
          new URLSearchParams(),
          new URLSearchParams(),
        ),
      ).toBe(false);
    });

    it("returns false when rillDefaults is undefined", () => {
      const yamlDefaults = new URLSearchParams("tr=P7D");
      expect(
        isViewingDefaults(
          new URLSearchParams(),
          yamlDefaults,
          undefined,
          new URLSearchParams("tr=P7D"),
        ),
      ).toBe(false);
    });

    it("handles YAML defaults that match rill defaults (minimal customization)", () => {
      // If YAML defaults = rill defaults, the browser URL should be empty
      const spec = makeExploreSpec({});
      const yamlDefaults = getYamlDefaultUrlParams(spec);
      const rillDefaults = getRillDefaultUrlParams(spec);

      // Empty browser URL = at rill defaults = at YAML defaults (since they're the same)
      const emptyBrowser = new URLSearchParams();
      const cleaned = cleanUrlParams(emptyBrowser, yamlDefaults);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, emptyBrowser),
      ).toBe(true);
    });
  });

  describe("with hardcoded URL strings", () => {
    // These tests use fully hardcoded URL strings to ensure deterministic behavior.
    // They don't depend on the YAML→state pipeline at all.

    it("detects viewing defaults with matching significant params", () => {
      // Scenario: YAML has tr and compare_tr different from rill defaults.
      // Rill defaults have tr=P7D, compare_tr=, grain=day
      // YAML defaults have tr=P14D, compare_tr=rill-PM, grain=day, f=, measures=*, dims=*
      // Browser URL (cleaned by DashboardStateSync): tr=P14D&compare_tr=rill-PM&grain=day
      //   grain is kept because tr differs from rill default
      const yamlDefaults = new URLSearchParams(
        "view=explore&tr=P14D&compare_tr=rill-PM&grain=day&f=&measures=*&dims=*&compare_dim=",
      );
      const rillDefaults = new URLSearchParams(
        "view=explore&tr=P7D&compare_tr=&grain=day&f=&measures=*&dims=*&compare_dim=",
      );
      const browserUrl = new URLSearchParams(
        "tr=P14D&compare_tr=rill-PM&grain=day",
      );
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      expect(cleaned.size).toBe(0);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(true);
    });

    it("detects NOT viewing defaults when filter is missing from URL", () => {
      // Scenario: YAML has a filter, browser URL doesn't
      const yamlDefaults = new URLSearchParams(
        "view=explore&tr=P14D&compare_tr=rill-PM&grain=day&f=customer+IN+%28%27Airtable%27%29&measures=*&dims=*&compare_dim=",
      );
      const rillDefaults = new URLSearchParams(
        "view=explore&tr=P7D&compare_tr=&grain=day&f=&measures=*&dims=*&compare_dim=",
      );
      // grain is present because tr differs from rill, but f is missing
      const browserUrl = new URLSearchParams(
        "tr=P14D&compare_tr=rill-PM&grain=day",
      );
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      // Forward check passes (all browser params match yaml)
      expect(cleaned.size).toBe(0);
      // But reverse check should fail (filter missing)
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(false);
    });

    it("detects viewing defaults when filter IS in the URL", () => {
      const yamlDefaults = new URLSearchParams(
        "view=explore&tr=P14D&compare_tr=rill-PM&grain=day&f=customer+IN+%28%27Airtable%27%29&measures=*&dims=*&compare_dim=",
      );
      const rillDefaults = new URLSearchParams(
        "view=explore&tr=P7D&compare_tr=&grain=day&f=&measures=*&dims=*&compare_dim=",
      );
      const browserUrl = new URLSearchParams(
        "tr=P14D&compare_tr=rill-PM&grain=day&f=customer+IN+%28%27Airtable%27%29",
      );
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      expect(cleaned.size).toBe(0);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(true);
    });

    it("handles filter normalization (array brackets)", () => {
      // YAML default uses array syntax from Go backend: IN (['Airtable'])
      // Browser URL uses plain syntax: IN ('Airtable')
      const yamlDefaults = new URLSearchParams(
        "view=explore&tr=P14D&grain=day&f=customer+IN+%28%5B%27Airtable%27%5D%29&measures=*&dims=*&compare_tr=&compare_dim=",
      );
      const rillDefaults = new URLSearchParams(
        "view=explore&tr=P7D&grain=day&f=&measures=*&dims=*&compare_tr=&compare_dim=",
      );
      const browserUrl = new URLSearchParams(
        "tr=P14D&grain=day&f=customer+IN+%28%27Airtable%27%29",
      );
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      expect(cleaned.size).toBe(0);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(true);
    });

    it("returns false when YAML has custom measures and URL is missing them", () => {
      // YAML specifies a subset of measures, rill defaults have all measures
      const yamlDefaults = new URLSearchParams(
        "view=explore&tr=P7D&grain=day&f=&measures=total_cost%2Ctotal_revenue&dims=*&compare_tr=&compare_dim=",
      );
      const rillDefaults = new URLSearchParams(
        "view=explore&tr=P7D&grain=day&f=&measures=*&dims=*&compare_tr=&compare_dim=",
      );
      // Browser URL has no measures param → implicitly rill default (*)
      const browserUrl = new URLSearchParams("");
      const cleaned = cleanUrlParams(browserUrl, yamlDefaults);
      expect(
        isViewingDefaults(cleaned, yamlDefaults, rillDefaults, browserUrl),
      ).toBe(false);
    });
  });
});
