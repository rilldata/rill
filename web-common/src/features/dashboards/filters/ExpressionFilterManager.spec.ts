import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters-v2/DimensionFilterManager.svelte.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters-v2/MeasureFilterManager.svelte.ts";
import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_COUNTRY_DIMENSION,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  createInEffectRoot,
  createTestMetricsViewsProvider,
  renderInRuntimeContext,
  useMetricsViewMocks,
  waitForMetricsViewSpecs,
} from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { flushSync } from "svelte";
import { get } from "svelte/store";
import {
  afterAll,
  afterEach,
  beforeAll,
  describe,
  expect,
  it,
  vi,
} from "vitest";

// ---------------------------------------------------------------------------
// Test metrics views
//
// AdBids and its mirror share `publisher` and `impressions`. Everything else is
// defined by exactly one of them, which is what makes the per-metrics-view
// behaviour observable:
//
//   publisher       both        domain          AdBids only
//   impressions     both        bid_price       AdBids only
//                               country         mirror only
//                               publisher_count mirror only
// ---------------------------------------------------------------------------

const AD_BIDS_MIRROR_METRICS_NAME = "AdBids_mirror_metrics";

useMetricsViewMocks({
  [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_INIT,
  [AD_BIDS_MIRROR_METRICS_NAME]: {
    ...AD_BIDS_METRICS_INIT,
    dimensions: [
      { name: AD_BIDS_PUBLISHER_DIMENSION },
      { name: AD_BIDS_COUNTRY_DIMENSION },
    ],
    measures: [
      { name: AD_BIDS_IMPRESSIONS_MEASURE, expression: "count(*)" },
      {
        name: AD_BIDS_PUBLISHER_COUNT_MEASURE,
        expression: "count_distinct(publisher)",
      },
    ],
  },
});

// The specs are read-only, so a single provider serves every test in this file.
// Each test still gets its own ExpressionFilterManager.
let metricsViewsProvider: MetricsViewsProvider;
let destroyProvider: () => void;
const cleanups: (() => void)[] = [];

beforeAll(async () => {
  const provider = await createTestMetricsViewsProvider([
    AD_BIDS_METRICS_NAME,
    AD_BIDS_MIRROR_METRICS_NAME,
  ]);
  metricsViewsProvider = provider.value;
  destroyProvider = provider.destroy;
});

afterEach(() => {
  cleanups.splice(0).forEach((destroy) => destroy());
});

afterAll(() => destroyProvider());

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** A real filter manager over the two metrics views above, torn down after the test. */
function createFilterManager(
  yamlConfigProvider: YAMLConfigProvider = new YAMLConfigProvider(),
) {
  const { value, destroy } = createInEffectRoot(
    () => new ExpressionFilterManager(metricsViewsProvider, yamlConfigProvider),
  );
  cleanups.push(destroy);
  return value;
}

/** The `f.<metricsView>` params, one filter string per metrics view. */
function perMetricsViewParams(filterByMetricsView: Record<string, string>) {
  const searchParams = new URLSearchParams();
  for (const [metricsViewName, filter] of Object.entries(filterByMetricsView)) {
    searchParams.set(
      `${ExploreStateURLParams.Filters}.${metricsViewName}`,
      filter,
    );
  }
  return searchParams;
}

/** The singular `f` param, which every metrics view falls back to. */
function sharedParam(filter: string) {
  return new URLSearchParams({ [ExploreStateURLParams.Filters]: filter });
}

function names(managers: { name: string }[]) {
  return managers.map((manager) => manager.name);
}

function dimensionManagersOf(
  filterManager: ExpressionFilterManager,
  metricsViewName: string,
) {
  return filterManager.managerByMetricsView[metricsViewName].managers
    .dimensionManagers;
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("setUrlParams", () => {
  it("builds a dimension chip from a per metrics view param", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual([
      "Google",
    ]);
    expect(filterManager.filterManagers.dimensions[0].mode).toBe(
      DimensionFilterMode.Select,
    );
    expect(filterManager.filterManagers.measures).toEqual([]);
    expect(filterManager.isComplexFilter).toBe(false);
  });

  it("builds a measure chip from a having param", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} gt 10)`,
      }),
    );

    expect(names(filterManager.filterManagers.measures)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
    const measureManager = filterManager.filterManagers.measures[0];
    expect(measureManager.dimension).toBe(AD_BIDS_PUBLISHER_DIMENSION);
    expect(measureManager.operation).toBe(MeasureFilterOperation.GreaterThan);
    expect(measureManager.value1).toBe("10");
    expect(filterManager.filterManagers.dimensions).toEqual([]);
  });

  it("reads an in-list filter back as in-list mode", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN LIST ('Google','Facebook')`,
      }),
    );

    expect(filterManager.filterManagers.dimensions[0].mode).toBe(
      DimensionFilterMode.InList,
    );
    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual([
      "Google",
      "Facebook",
    ]);
  });

  it("applies the singular param to every metrics view and shares one manager", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`),
    );

    expect(Object.keys(filterManager.managerByMetricsView)).toEqual([
      AD_BIDS_METRICS_NAME,
      AD_BIDS_MIRROR_METRICS_NAME,
    ]);
    // One chip, backed by a single manager instance, so editing it updates the
    // filter for both metrics views at once.
    expect(filterManager.filterManagers.dimensions).toHaveLength(1);
    expect(dimensionManagersOf(filterManager, AD_BIDS_METRICS_NAME)[0]).toBe(
      dimensionManagersOf(filterManager, AD_BIDS_MIRROR_METRICS_NAME)[0],
    );
  });

  it("prefers the per metrics view param over the singular one", () => {
    const filterManager = createFilterManager();

    const searchParams = sharedParam(
      `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
    );
    searchParams.set(
      `${ExploreStateURLParams.Filters}.${AD_BIDS_MIRROR_METRICS_NAME}`,
      `${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`,
    );
    filterManager.setUrlParams(searchParams);

    expect(
      names(dimensionManagersOf(filterManager, AD_BIDS_METRICS_NAME)),
    ).toEqual([AD_BIDS_PUBLISHER_DIMENSION]);
    expect(
      names(dimensionManagersOf(filterManager, AD_BIDS_MIRROR_METRICS_NAME)),
    ).toEqual([AD_BIDS_COUNTRY_DIMENSION]);
    // Chips are the union across metrics views, sorted by name.
    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_COUNTRY_DIMENSION,
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
  });

  it("drops filters on identifiers the metrics view does not define", () => {
    const filterManager = createFilterManager();

    // `country` belongs to the mirror only, so AdBids cannot filter on it.
    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`,
      }),
    );

    expect(dimensionManagersOf(filterManager, AD_BIDS_METRICS_NAME)).toEqual(
      [],
    );
    expect(filterManager.filterManagers.dimensions).toEqual([]);
  });

  it("keeps the managers when the params do not change", () => {
    const filterManager = createFilterManager();
    const filter = `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`;

    filterManager.setUrlParams(
      perMetricsViewParams({ [AD_BIDS_METRICS_NAME]: filter }),
    );
    const dimensionManager = filterManager.filterManagers.dimensions[0];

    filterManager.setUrlParams(
      perMetricsViewParams({ [AD_BIDS_METRICS_NAME]: filter }),
    );

    expect(filterManager.filterManagers.dimensions[0]).toBe(dimensionManager);
  });

  it("flags an OR filter as complex", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') OR ${AD_BIDS_DOMAIN_DIMENSION} IN ('google.com')`,
      }),
    );

    expect(filterManager.isComplexFilter).toBe(true);
    expect(
      filterManager.managerByMetricsView[AD_BIDS_METRICS_NAME].isComplexFilter,
    ).toBe(true);
  });

  it("flags two filters on the same dimension as complex", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') AND ${AD_BIDS_PUBLISHER_DIMENSION} NIN ('Facebook')`,
      }),
    );

    expect(
      filterManager.managerByMetricsView[AD_BIDS_METRICS_NAME].isComplexFilter,
    ).toBe(true);
  });

  it("clears the temporary filter name", () => {
    const filterManager = createFilterManager();

    filterManager.addNewFilter(AD_BIDS_DOMAIN_DIMENSION);
    expect(filterManager.temporaryFilterName).toBe(AD_BIDS_DOMAIN_DIMENSION);

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    expect(filterManager.temporaryFilterName).toBeUndefined();
    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
  });
});

describe("setExprForMetricsView / setParamForMetricsView", () => {
  it("sets the filter for one metrics view without touching the others", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`),
    );
    filterManager.setParamForMetricsView(
      AD_BIDS_MIRROR_METRICS_NAME,
      `${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`,
    );

    expect(
      names(dimensionManagersOf(filterManager, AD_BIDS_METRICS_NAME)),
    ).toEqual([AD_BIDS_PUBLISHER_DIMENSION]);
    expect(
      names(dimensionManagersOf(filterManager, AD_BIDS_MIRROR_METRICS_NAME)),
    ).toEqual([AD_BIDS_COUNTRY_DIMENSION]);
  });

  it("converts an expression to a param", () => {
    const filterManager = createFilterManager();

    filterManager.setExprForMetricsView(
      AD_BIDS_METRICS_NAME,
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google", "Facebook"]),
      ]),
    );

    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual([
      "Google",
      "Facebook",
    ]);
  });

  it("clears the filter for a metrics view when given no expression", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`),
    );
    filterManager.setExprForMetricsView(AD_BIDS_METRICS_NAME, undefined);

    expect(
      filterManager.managerByMetricsView[AD_BIDS_METRICS_NAME].expr,
    ).toBeUndefined();
    expect(
      names(dimensionManagersOf(filterManager, AD_BIDS_MIRROR_METRICS_NAME)),
    ).toEqual([AD_BIDS_PUBLISHER_DIMENSION]);
  });
});

describe("chip order", () => {
  it("leads with required filters, then pinned, then the param filters by name", () => {
    const yamlConfigProvider = new YAMLConfigProvider();
    yamlConfigProvider.requiredFilters = { [AD_BIDS_DOMAIN_DIMENSION]: true };
    yamlConfigProvider.pinnedFilters = { [AD_BIDS_PUBLISHER_DIMENSION]: true };
    const filterManager = createFilterManager(yamlConfigProvider);

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') AND ${AD_BIDS_DOMAIN_DIMENSION} IN ('google.com')`,
        [AD_BIDS_MIRROR_METRICS_NAME]: `${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`,
      }),
    );

    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_DOMAIN_DIMENSION,
      AD_BIDS_PUBLISHER_DIMENSION,
      AD_BIDS_COUNTRY_DIMENSION,
    ]);
  });

  it("gives required and pinned filters a chip without a value", () => {
    const yamlConfigProvider = new YAMLConfigProvider();
    yamlConfigProvider.requiredFilters = { [AD_BIDS_DOMAIN_DIMENSION]: true };
    yamlConfigProvider.pinnedFilters = { [AD_BIDS_IMPRESSIONS_MEASURE]: true };
    const filterManager = createFilterManager(yamlConfigProvider);

    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_DOMAIN_DIMENSION,
    ]);
    expect(filterManager.filterManagers.dimensions[0].expr).toBeUndefined();
    expect(names(filterManager.filterManagers.measures)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
    expect(filterManager.filterManagers.measures[0].expr).toBeUndefined();
  });

  it("skips required filters that no metrics view defines", () => {
    const yamlConfigProvider = new YAMLConfigProvider();
    yamlConfigProvider.requiredFilters = { not_a_field: true };
    const filterManager = createFilterManager(yamlConfigProvider);

    expect(filterManager.filterManagers.dimensions).toEqual([]);
    expect(filterManager.filterManagers.measures).toEqual([]);
  });

  it("sorts the param measures by name", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} gt 10) AND ${AD_BIDS_DOMAIN_DIMENSION} having (${AD_BIDS_BID_PRICE_MEASURE} lt 2)`,
      }),
    );

    expect(names(filterManager.filterManagers.measures)).toEqual([
      AD_BIDS_BID_PRICE_MEASURE,
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
  });

  it("keys filterManagersMap by chip name", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') AND ${AD_BIDS_DOMAIN_DIMENSION} having (${AD_BIDS_BID_PRICE_MEASURE} lt 2)`,
      }),
    );

    expect(Object.keys(filterManager.filterManagersMap).sort()).toEqual([
      AD_BIDS_BID_PRICE_MEASURE,
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
    expect(
      filterManager.filterManagersMap[AD_BIDS_PUBLISHER_DIMENSION],
    ).toBeInstanceOf(DimensionFilterManager);
    expect(
      filterManager.filterManagersMap[AD_BIDS_BID_PRICE_MEASURE],
    ).toBeInstanceOf(MeasureFilterManager);
  });
});

describe("addNewFilter", () => {
  it("opens a chip for a dimension that has no value yet", () => {
    const filterManager = createFilterManager();

    filterManager.addNewFilter(AD_BIDS_DOMAIN_DIMENSION);

    expect(filterManager.temporaryFilterName).toBe(AD_BIDS_DOMAIN_DIMENSION);
    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_DOMAIN_DIMENSION,
    ]);
    expect(filterManager.filterManagers.dimensions[0].expr).toBeUndefined();
  });

  it("opens a chip for a measure that has no value yet", () => {
    const filterManager = createFilterManager();

    filterManager.addNewFilter(AD_BIDS_BID_PRICE_MEASURE);

    expect(names(filterManager.filterManagers.measures)).toEqual([
      AD_BIDS_BID_PRICE_MEASURE,
    ]);
    expect(filterManager.filterManagers.dimensions).toEqual([]);
  });

  it("ignores a name no metrics view defines", () => {
    const filterManager = createFilterManager();

    filterManager.addNewFilter("not_a_field");

    expect(filterManager.temporaryFilterName).toBeUndefined();
    expect(filterManager.filterManagers.dimensions).toEqual([]);
    expect(filterManager.filterManagers.measures).toEqual([]);
  });

  it("does not duplicate a chip the param already covers", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );
    filterManager.addNewFilter(AD_BIDS_PUBLISHER_DIMENSION);

    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual([
      "Google",
    ]);
  });
});

describe("dimensionFilterAction", () => {
  it("creates and keeps a manager for a dimension the param does not cover", () => {
    const filterManager = createFilterManager();

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.setSelectedValues(["Google"], false),
    );

    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual([
      "Google",
    ]);
    expect(
      filterManager.filterManagersMap[AD_BIDS_PUBLISHER_DIMENSION],
    ).toBeInstanceOf(DimensionFilterManager);
  });

  it("drops the new manager when the callback leaves it without an expression", () => {
    const filterManager = createFilterManager();

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.clear(),
    );

    expect(filterManager.filterManagers.dimensions).toEqual([]);
  });

  it("reuses the manager the param created", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );
    const existing = filterManager.filterManagers.dimensions[0];

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => {
        expect(manager).toBe(existing);
        manager.toggleValue("Facebook", false);
      },
    );

    expect(existing.selectedValues).toEqual(["Google", "Facebook"]);
    expect(filterManager.filterManagers.dimensions).toHaveLength(1);
  });

  it("returns the callback result", () => {
    const filterManager = createFilterManager();

    // The callback return type is `any`, hence the cast.
    const result = filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.selectedValues.length,
    ) as number;

    expect(result).toBe(0);
  });

  it("does nothing for a measure or an unknown name", () => {
    const filterManager = createFilterManager();
    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} gt 10)`,
      }),
    );
    const callback = vi.fn();

    expect(
      filterManager.dimensionFilterAction(
        AD_BIDS_IMPRESSIONS_MEASURE,
        callback,
      ),
    ).toBeUndefined();
    expect(
      filterManager.dimensionFilterAction("not_a_field", callback),
    ).toBeUndefined();
    expect(callback).not.toHaveBeenCalled();
  });

  it("set => unset => set filters should add filter", () => {
    const filterManager = createFilterManager();

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.toggleValue("Google", false),
    );
    expect(filterManager.filterManagers.dimensions[0]).not.toBeUndefined();
    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual([
      "Google",
    ]);

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.toggleValue("Google", false),
    );
    expect(filterManager.filterManagers.dimensions[0]).not.toBeUndefined();
    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual(
      [],
    );

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.toggleValue("Google", false),
    );
    expect(filterManager.filterManagers.dimensions[0]).not.toBeUndefined();
    expect(filterManager.filterManagers.dimensions[0].selectedValues).toEqual([
      "Google",
    ]);
  });
});

describe("clear", () => {
  it("removes every chip and the temporary filter", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
        [AD_BIDS_MIRROR_METRICS_NAME]: `${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`,
      }),
    );
    filterManager.addNewFilter(AD_BIDS_DOMAIN_DIMENSION);

    filterManager.clear();

    expect(filterManager.temporaryFilterName).toBeUndefined();
    expect(
      Object.values(filterManager.managerByMetricsView).map((m) => m.expr),
    ).toEqual([undefined, undefined]);
    expect(filterManager.filterManagers.dimensions).toEqual([]);
    expect(filterManager.filterManagers.measures).toEqual([]);
  });

  it("keeps the required and pinned chips", () => {
    const yamlConfigProvider = new YAMLConfigProvider();
    yamlConfigProvider.requiredFilters = { [AD_BIDS_DOMAIN_DIMENSION]: true };
    const filterManager = createFilterManager(yamlConfigProvider);

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_DOMAIN_DIMENSION} IN ('google.com')`,
      }),
    );
    filterManager.clear();

    expect(names(filterManager.filterManagers.dimensions)).toEqual([
      AD_BIDS_DOMAIN_DIMENSION,
    ]);
    expect(filterManager.filterManagers.dimensions[0].expr).toBeUndefined();
  });
});

describe("getOtherDimensionsFilter", () => {
  function setupThreeDimensions() {
    const filterManager = createFilterManager();
    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') AND ${AD_BIDS_DOMAIN_DIMENSION} IN ('google.com')`,
        [AD_BIDS_MIRROR_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') AND ${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`,
      }),
    );
    return filterManager;
  }

  it("ands the other dimension filters that apply to the metrics view", () => {
    const filterManager = setupThreeDimensions();

    // `country` is a mirror-only dimension, so it is left out of the AdBids filter.
    expect(
      filterManager.getOtherDimensionsFilter(
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_METRICS_NAME,
      ),
    ).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["google.com"]),
      ]),
    );
  });

  it("only includes dimensions the given metrics view defines", () => {
    const filterManager = setupThreeDimensions();

    expect(
      filterManager.getOtherDimensionsFilter(
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_MIRROR_METRICS_NAME,
      ),
    ).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_COUNTRY_DIMENSION, ["US"]),
      ]),
    );
  });

  it("returns undefined when no other dimension is filtered", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    expect(
      filterManager.getOtherDimensionsFilter(
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_METRICS_NAME,
      ),
    ).toBeUndefined();
  });
});

describe("manager factories", () => {
  it("creates a dimension manager for a dimension name", () => {
    const filterManager = createFilterManager();

    const manager = filterManager.createDimensionOrMeasureManagerForName(
      AD_BIDS_PUBLISHER_DIMENSION,
    );

    expect(manager).toBeInstanceOf(DimensionFilterManager);
    expect(manager?.name).toBe(AD_BIDS_PUBLISHER_DIMENSION);
    expect(manager?.label).toBe(AD_BIDS_PUBLISHER_DIMENSION);
    expect(manager?.expr).toBeUndefined();
  });

  it("creates a measure manager for a measure name", () => {
    const filterManager = createFilterManager();

    const manager = filterManager.createDimensionOrMeasureManagerForName(
      AD_BIDS_BID_PRICE_MEASURE,
    );

    expect(manager).toBeInstanceOf(MeasureFilterManager);
    expect(manager?.name).toBe(AD_BIDS_BID_PRICE_MEASURE);
  });

  it("returns undefined for a name no metrics view defines", () => {
    const filterManager = createFilterManager();

    expect(
      filterManager.createDimensionOrMeasureManagerForName("not_a_field"),
    ).toBeUndefined();
  });

  it("creates a dimension manager from an in expression", () => {
    const filterManager = createFilterManager();

    const manager = filterManager.createDimensionOrMeasureManagerForExpr(
      createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
    );

    expect(manager).toBeInstanceOf(DimensionFilterManager);
    expect((manager as DimensionFilterManager).selectedValues).toEqual([
      "Google",
    ]);
    expect((manager as DimensionFilterManager).mode).toBe(
      DimensionFilterMode.Select,
    );
  });

  it("honours the in-list dimensions when creating from an expression", () => {
    const filterManager = createFilterManager();

    const manager = filterManager.createDimensionOrMeasureManagerForExpr(
      createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      [AD_BIDS_PUBLISHER_DIMENSION],
    );

    expect((manager as DimensionFilterManager).mode).toBe(
      DimensionFilterMode.InList,
    );
  });

  it("creates a measure manager from a subquery expression", () => {
    const filterManager = createFilterManager();

    const manager = filterManager.createDimensionOrMeasureManagerForExpr(
      createSubQueryExpression(
        AD_BIDS_PUBLISHER_DIMENSION,
        [AD_BIDS_IMPRESSIONS_MEASURE],
        createBinaryExpression(
          AD_BIDS_IMPRESSIONS_MEASURE,
          V1Operation.OPERATION_GT,
          10,
        ),
      ),
    );

    expect(manager).toBeInstanceOf(MeasureFilterManager);
    expect((manager as MeasureFilterManager).dimension).toBe(
      AD_BIDS_PUBLISHER_DIMENSION,
    );
    expect((manager as MeasureFilterManager).value1).toBe("10");
  });

  it("returns undefined when the expression is keyed off an unknown dimension", () => {
    const filterManager = createFilterManager();

    expect(
      filterManager.createDimensionOrMeasureManagerForExpr(
        createInExpression("not_a_dimension", ["Google"]),
      ),
    ).toBeUndefined();
  });
});

describe("createLocalFilterStore", () => {
  it("scopes a new manager to a single metrics view and shares the yaml config", async () => {
    const yamlConfigProvider = new YAMLConfigProvider();
    // The local store builds its own MetricsViewsProvider, which calls createQuery,
    // so it has to be created inside a component rather than a bare effect root.
    const rendered = renderInRuntimeContext(({ runtimeClient }) => {
      const provider = new MetricsViewsProvider(runtimeClient, [
        AD_BIDS_METRICS_NAME,
        AD_BIDS_MIRROR_METRICS_NAME,
      ]);
      const filterManager = new ExpressionFilterManager(
        provider,
        yamlConfigProvider,
      );
      return {
        provider,
        filterManager,
        local: filterManager.createLocalFilterStore(AD_BIDS_METRICS_NAME),
      };
    });
    cleanups.push(() => {
      rendered.value.local.metricsViewsProvider.cleanup();
      rendered.value.provider.cleanup();
      rendered.destroy();
    });

    const { local } = rendered.value;
    await waitForMetricsViewSpecs(local.metricsViewsProvider, [
      AD_BIDS_METRICS_NAME,
    ]);

    expect(local.metricsViewsProvider.metricsViewNames).toEqual([
      AD_BIDS_METRICS_NAME,
    ]);
    expect(local.yamlConfigProvider).toBe(yamlConfigProvider);

    // The mirror-only dimension is out of scope for the local store.
    local.setUrlParams(sharedParam(`${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`));
    expect(local.filterManagers.dimensions).toEqual([]);

    local.setUrlParams(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`),
    );
    expect(names(local.filterManagers.dimensions)).toEqual([
      AD_BIDS_PUBLISHER_DIMENSION,
    ]);
    // The parent manager is untouched.
    expect(rendered.value.filterManager.filterManagers.dimensions).toEqual([]);
  });
});

describe("exprByMetricsView", () => {
  it("builds an expression per metrics view", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') AND ${AD_BIDS_DOMAIN_DIMENSION} IN ('google.com')`,
        [AD_BIDS_MIRROR_METRICS_NAME]: `${AD_BIDS_COUNTRY_DIMENSION} IN ('US')`,
      }),
    );

    expect(filterManager.exprByMetricsView).toEqual({
      [AD_BIDS_METRICS_NAME]: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
        createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["google.com"]),
      ]),
      [AD_BIDS_MIRROR_METRICS_NAME]: createAndExpression([
        createInExpression(AD_BIDS_COUNTRY_DIMENSION, ["US"]),
      ]),
    });
    expect(filterManager.hasSomeFilter).toBe(true);
  });

  it("omits metrics views with no filter", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_DOMAIN_DIMENSION} IN ('google.com')`,
      }),
    );

    expect(Object.keys(filterManager.exprByMetricsView)).toEqual([
      AD_BIDS_METRICS_NAME,
    ]);
  });

  it("is empty before any filter is set", () => {
    const filterManager = createFilterManager();

    expect(filterManager.exprByMetricsView).toEqual({});
    expect(filterManager.hasSomeFilter).toBe(false);
  });

  it("picks up a filter added outside the params", async () => {
    const filterManager = createFilterManager();

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.setSelectedValues(["Google"], false),
    );

    // `publisher` belongs to both metrics views, so both get the filter.
    expect(filterManager.exprByMetricsView).toEqual({
      [AD_BIDS_METRICS_NAME]: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      ]),
      [AD_BIDS_MIRROR_METRICS_NAME]: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      ]),
    });
  });

  it("mirrors exprByMetricsView into the svelte 4 store", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_DOMAIN_DIMENSION} IN ('google.com')`,
      }),
    );
    flushSync();

    expect(get(filterManager.exprByMetricsViewStore)).toEqual(
      filterManager.exprByMetricsView,
    );
  });
});
