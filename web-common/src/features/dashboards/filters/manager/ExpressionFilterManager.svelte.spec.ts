import { describe, expect, it } from "vitest";
import { ExpressionFilterManager } from "./ExpressionFilterManager.svelte.ts";
import { DimensionFilterManager } from "./DimensionFilterManager.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";

const MetricsViewName = "mv";

/** Stands in for the specs a MetricsViewsProvider exposes, so they can arrive mid-test. */
class TestMetricsViewsProvider {
  public dimensionSpecs = $state<
    Record<string, Record<string, MetricsViewSpecDimension>>
  >({});
  public measureSpecs = $state<
    Record<string, Record<string, MetricsViewSpecMeasure>>
  >({});

  public loadSpecs() {
    this.dimensionSpecs = {
      country: { [MetricsViewName]: { name: "country" } },
      state: { [MetricsViewName]: { name: "state" } },
    };
    this.measureSpecs = {
      revenue: { [MetricsViewName]: { name: "revenue" } },
    };
  }
}

function createManager() {
  const metricsViewsProvider = new TestMetricsViewsProvider();
  const yamlConfigProvider = new YAMLConfigProvider(true);
  const filterManager = new ExpressionFilterManager(
    metricsViewsProvider as unknown as MetricsViewsProvider,
    yamlConfigProvider,
  );
  return { filterManager, metricsViewsProvider, yamlConfigProvider };
}

function chipNames(filterManager: ExpressionFilterManager) {
  return [
    ...filterManager.dimensionFilterManagers.map((dfm) => dfm.name),
    ...filterManager.measureFilterManagers.map((mfm) => mfm.name),
  ];
}

describe("ExpressionFilterManager", () => {
  it("builds dimension and measure managers from the param", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider } = createManager();
      metricsViewsProvider.loadSpecs();

      filterManager.setExprParam(
        "country IN ('US','IN') AND state having (revenue GTE 10)",
      );

      expect(
        filterManager.dimensionFilterManagers.map((dfm) => [
          dfm.name,
          dfm.selectedValues,
        ]),
      ).toEqual([["country", ["US", "IN"]]]);
      expect(
        filterManager.measureFilterManagers.map((mfm) => [
          mfm.name,
          mfm.dimension,
          mfm.value1,
        ]),
      ).toEqual([["revenue", "state", "10"]]);
      expect(filterManager.isComplexFilter).toBe(false);
    });
    cleanup();
  });

  it("rebuilds once the metrics view specs arrive", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider } = createManager();

      // The param is set before ListResources resolves, so nothing can be resolved yet.
      filterManager.setExprParam("country IN ('US')");
      expect(chipNames(filterManager)).toEqual([]);

      metricsViewsProvider.loadSpecs();
      expect(chipNames(filterManager)).toEqual(["country"]);
    });
    cleanup();
  });

  it("leads with required filters, then pinned ones, and keeps a chip for both without a value", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider, yamlConfigProvider } =
        createManager();
      metricsViewsProvider.loadSpecs();
      yamlConfigProvider.togglePinnedFilter("country");
      yamlConfigProvider.toggleRequiredFilter("state");

      filterManager.setExprParam("country IN ('US')");

      // "state" is required and has no value, so it gets an empty chip ahead of the pinned one.
      expect(chipNames(filterManager)).toEqual(["state", "country"]);
      const [stateManager] = filterManager.dimensionFilterManagers;
      expect(stateManager.expr).toBeUndefined();
      // An empty chip contributes nothing to the expression.
      expect(convertExpressionToFilterParam(filterManager.expr)).toEqual(
        "country IN ('US')",
      );
    });
    cleanup();
  });

  it("adds and removes chips when the pinned config changes", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider, yamlConfigProvider } =
        createManager();
      metricsViewsProvider.loadSpecs();
      filterManager.setExprParam("country IN ('US')");

      yamlConfigProvider.togglePinnedFilter("revenue");
      expect(chipNames(filterManager)).toEqual(["country", "revenue"]);

      yamlConfigProvider.togglePinnedFilter("revenue");
      expect(chipNames(filterManager)).toEqual(["country"]);
    });
    cleanup();
  });

  it("keeps a value applied to a required chip while the pinned config changes with it", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider, yamlConfigProvider } =
        createManager();
      metricsViewsProvider.loadSpecs();
      yamlConfigProvider.toggleRequiredFilter("country");
      filterManager.setExprParam("");

      const [countryManager] = filterManager.dimensionFilterManagers;
      // The filter chip persists pinned/required together with the value it was given.
      yamlConfigProvider.togglePinnedFilter("country");
      countryManager.setSelectedValues(["US"], false);

      expect(convertExpressionToFilterParam(filterManager.expr)).toEqual(
        "country IN ('US')",
      );
    });
    cleanup();
  });

  it("does not restore a stale value after a required filter is cleared", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider, yamlConfigProvider } =
        createManager();
      metricsViewsProvider.loadSpecs();
      yamlConfigProvider.toggleRequiredFilter("country");
      filterManager.setExprParam("");

      filterManager.dimensionFilterManagers[0].setSelectedValues(["US"], false);
      // The value reaches the url, so the param takes over from here.
      filterManager.setExprParam(
        convertExpressionToFilterParam(filterManager.expr),
      );
      expect(filterManager.dimensionFilterManagers[0].selectedValues).toEqual([
        "US",
      ]);

      filterManager.dimensionFilterManagers[0].clear();
      filterManager.setExprParam(
        convertExpressionToFilterParam(filterManager.expr),
      );

      // The chip stays because the filter is required, but without its old value.
      expect(chipNames(filterManager)).toEqual(["country"]);
      expect(filterManager.dimensionFilterManagers[0].selectedValues).toEqual(
        [],
      );
    });
    cleanup();
  });

  it("keeps a temporary filter until the param covers it", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider } = createManager();
      metricsViewsProvider.loadSpecs();
      filterManager.setExprParam("country IN ('US')");

      filterManager.addTemporaryFilter("state");
      expect(chipNames(filterManager)).toEqual(["country", "state"]);
      expect(filterManager.temporaryFilterName).toEqual("state");

      // An unrelated url change keeps the chip, since the param still does not cover it.
      filterManager.setExprParam("country IN ('US')");
      expect(chipNames(filterManager)).toEqual(["country", "state"]);

      filterManager.setExprParam("country IN ('US') AND state IN ('CA')");
      expect(chipNames(filterManager)).toEqual(["country", "state"]);
      expect(filterManager.temporaryFilterName).toBeUndefined();
    });
    cleanup();
  });

  it("keeps a chip for a filter created by a dimension filter action", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider } = createManager();
      metricsViewsProvider.loadSpecs();
      filterManager.setExprParam("country IN ('US')");

      filterManager.dimensionFilterAction("state", (dimensionFilterManager) =>
        dimensionFilterManager.toggleValue("CA", false),
      );

      expect(chipNames(filterManager)).toEqual(["country", "state"]);
      expect(convertExpressionToFilterParam(filterManager.expr)).toEqual(
        "country IN ('US') AND state IN ('CA')",
      );
    });
    cleanup();
  });

  it("routes a dimension filter action to the manager the param already has", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider } = createManager();
      metricsViewsProvider.loadSpecs();
      filterManager.setExprParam("country IN ('US')");

      filterManager.dimensionFilterAction("country", (dimensionFilterManager) =>
        dimensionFilterManager.toggleValue("IN", false),
      );

      expect(chipNames(filterManager)).toEqual(["country"]);
      expect(convertExpressionToFilterParam(filterManager.expr)).toEqual(
        "country IN ('US','IN')",
      );
    });
    cleanup();
  });

  it("clear drops every filter but keeps the required and pinned chips", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider, yamlConfigProvider } =
        createManager();
      metricsViewsProvider.loadSpecs();
      yamlConfigProvider.togglePinnedFilter("state");
      filterManager.setExprParam("country IN ('US') AND state IN ('CA')");
      filterManager.addTemporaryFilter("revenue");

      filterManager.clear();

      expect(chipNames(filterManager)).toEqual(["state"]);
      expect(filterManager.dimensionFilterManagers[0].expr).toBeUndefined();
      expect(filterManager.temporaryFilterName).toBeUndefined();
    });
    cleanup();
  });

  it("has no chips for a filter it cannot represent", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider } = createManager();
      metricsViewsProvider.loadSpecs();

      filterManager.setExprParam("country IN ('US') OR state IN ('CA')");

      expect(filterManager.isComplexFilter).toBe(true);
      expect(chipNames(filterManager)).toEqual([]);
    });
    cleanup();
  });

  it("skips filters no metrics view defines", () => {
    const cleanup = $effect.root(() => {
      const { filterManager, metricsViewsProvider, yamlConfigProvider } =
        createManager();
      metricsViewsProvider.loadSpecs();
      yamlConfigProvider.togglePinnedFilter("unknown_dimension");

      filterManager.setExprParam("country IN ('US') AND removed IN ('CA')");

      expect(chipNames(filterManager)).toEqual(["country"]);
      expect(
        filterManager.dimensionFilterManagers[0] instanceof
          DimensionFilterManager,
      ).toBe(true);
    });
    cleanup();
  });
});
