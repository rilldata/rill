import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  createAndExpression,
  createInExpression,
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
import { flushSync } from "svelte";
import { fromStore, get, writable } from "svelte/store";
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

/**
 * A filter manager wired the way a canvas wires one: seeded from the params its entity holds,
 * writing back to the page url. The two sources are separate on purpose, since the seed trails the
 * page url by the redirect handling in `CanvasEntity.onUrlChange`.
 */
function createUrlSyncedFilterManager() {
  const seedParams = writable(new URLSearchParams());
  const url = writable(new URL("http://localhost/canvas"));
  const navigations: URL[] = [];
  let filterManager!: ExpressionFilterManager;
  let disposeSeed!: () => void;

  const { destroy } = createInEffectRoot(() => {
    filterManager = new ExpressionFilterManager(
      metricsViewsProvider,
      new YAMLConfigProvider(),
    );
    disposeSeed = filterManager.seedFromSearchParams(seedParams);
    const currentUrl = fromStore(url);
    filterManager.writeBackToUrl(
      () => currentUrl.current,
      (navigated) => navigations.push(navigated),
    );
  });
  cleanups.push(() => {
    disposeSeed();
    destroy();
  });

  /** Both sources at once, as they move for a canvas that is not mid navigation. */
  function setUrl(searchParams: URLSearchParams) {
    url.set(new URL(`http://localhost/canvas?${searchParams.toString()}`));
    seedParams.set(searchParams);
    flushSync();
  }

  return { filterManager, navigations, seedParams, url, setUrl };
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

  it("flags a dimension filtered differently in each metrics view as complex", () => {
    const filterManager = createFilterManager();

    // A single chip can only show one of the two, and editing it would leave the mirror on
    // `Facebook`, so the whole filter falls back to the read only advanced filter.
    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
        [AD_BIDS_MIRROR_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook')`,
      }),
    );

    expect(filterManager.isComplexFilter).toBe(true);
    // Neither metrics view is complex on its own; each holds a single plain filter.
    expect(
      Object.values(filterManager.managerByMetricsView).some(
        (manager) => manager.isComplexFilter,
      ),
    ).toBe(false);
  });

  it("does not flag the same filter in every metrics view as complex", () => {
    const filterManager = createFilterManager();

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
        [AD_BIDS_MIRROR_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    expect(filterManager.isComplexFilter).toBe(false);
    expect(filterManager.filterManagers.dimensions).toHaveLength(1);
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

describe("applyFilterToParams", () => {
  it("writes a param per metrics view and drops the legacy singular one", () => {
    const filterManager = createFilterManager();
    filterManager.setUrlParams(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`),
    );

    // The singular param seeded every metrics view, so keeping it would be a second source of
    // truth: removing the chip could not delete it, and the next read would restore the filter.
    const searchParams = sharedParam(
      `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
    );
    filterManager.applyFilterToParams(searchParams);

    expect(searchParams.toString()).toEqual(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
        [AD_BIDS_MIRROR_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }).toString(),
    );
  });

  it("leaves nothing behind once the filter is removed", () => {
    const filterManager = createFilterManager();
    filterManager.setUrlParams(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`),
    );
    filterManager.filterManagers.dimensions[0].clear();

    const searchParams = sharedParam(
      `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
    );
    filterManager.applyFilterToParams(searchParams);

    expect(searchParams.toString()).toEqual("");
  });

  it("keeps every filter under the singular param when asked for one", () => {
    const filterManager = createFilterManager();
    filterManager.setUrlParams(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`),
    );

    const searchParams = new URLSearchParams();
    filterManager.applyFilterToParams(searchParams, true);

    expect(searchParams.toString()).toEqual(
      sharedParam(`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`).toString(),
    );
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

describe("filter-changed", () => {
  /** Records the source of every `filter-changed` the manager reports. */
  function recordSources(filterManager: ExpressionFilterManager) {
    const sources: (string | undefined)[] = [];
    cleanups.push(
      filterManager.on("filter-changed", ({ source }) => {
        sources.push(source);
      }),
    );
    return sources;
  }

  it("reports a chip edit with no source", () => {
    const filterManager = createFilterManager();
    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );
    const sources = recordSources(filterManager);

    // The chips hold the managers and mutate them directly, without going through the root.
    filterManager.filterManagers.dimensions[0].toggleValue("Facebook", false);

    expect(sources).toEqual([undefined]);
  });

  it("reports the source an action was applied with", () => {
    const filterManager = createFilterManager();
    const sources = recordSources(filterManager);

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.setSelectedValues(["Google"], false),
      "component-1",
    );
    filterManager.measureFilterAction(
      AD_BIDS_IMPRESSIONS_MEASURE,
      (manager) =>
        manager.setMeasureFilter(AD_BIDS_PUBLISHER_DIMENSION, {
          measure: AD_BIDS_IMPRESSIONS_MEASURE,
          type: MeasureFilterType.Value,
          operation: MeasureFilterOperation.GreaterThan,
          value1: "10",
          value2: "",
        }),
      "component-2",
    );

    expect(sources).toEqual(["component-1", "component-2"]);
  });

  it("does not carry the source past the action it was given to", () => {
    const filterManager = createFilterManager();
    const sources = recordSources(filterManager);

    filterManager.dimensionFilterAction(
      AD_BIDS_PUBLISHER_DIMENSION,
      (manager) => manager.setSelectedValues(["Google"], false),
      "component-1",
    );
    // Same manager, edited from the chip this time.
    filterManager.filterManagersMap[AD_BIDS_PUBLISHER_DIMENSION].clear();

    expect(sources).toEqual(["component-1", undefined]);
  });

  it("reports clear with no source", () => {
    const filterManager = createFilterManager();
    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );
    const sources = recordSources(filterManager);

    filterManager.clear();

    expect(sources).toEqual([undefined]);
  });

  it("stays quiet while the param is parsed into managers", () => {
    const filterManager = createFilterManager();
    const sources = recordSources(filterManager);

    filterManager.setUrlParams(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google') AND ${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} gt 10)`,
      }),
    );
    // Reading the chips rebuilds every manager from the param.
    expect(filterManager.filterManagers.dimensions).toHaveLength(1);
    expect(filterManager.filterManagers.measures).toHaveLength(1);
    flushSync();

    // A component that filters writes its change back to the url, which lands here.
    // Reporting the reparse would look like a second, sourceless change.
    expect(sources).toEqual([]);
  });
});

describe("seedFromSearchParams", () => {
  it("seeds the managers without a filter bar", () => {
    const { filterManager, seedParams } = createUrlSyncedFilterManager();

    seedParams.set(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );
    flushSync();

    expect(filterManager.exprByMetricsView[AD_BIDS_METRICS_NAME]).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      ]),
    );
  });

  it("stops seeding once disposed", () => {
    const { filterManager, seedParams } = createUrlSyncedFilterManager();
    const disposed = cleanups.pop() as () => void;

    disposed();
    seedParams.set(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );
    flushSync();

    expect(filterManager.hasSomeFilter).toBe(false);
  });
});

describe("writeBackToUrl", () => {
  it("leaves the url alone while it matches the managers", () => {
    const { navigations, setUrl } = createUrlSyncedFilterManager();

    setUrl(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    expect(navigations).toEqual([]);
  });

  it("writes a chip edit back", () => {
    const { filterManager, navigations, setUrl } =
      createUrlSyncedFilterManager();
    setUrl(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    filterManager.filterManagers.dimensions[0].setSelectedValues(
      ["Google", "Facebook"],
      false,
    );
    flushSync();

    expect(
      navigations.map((url) =>
        url.searchParams.get(
          `${ExploreStateURLParams.Filters}.${AD_BIDS_METRICS_NAME}`,
        ),
      ),
    ).toEqual([`${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google','Facebook')`]);
  });

  it("writes a clear back", () => {
    const { filterManager, navigations, setUrl } =
      createUrlSyncedFilterManager();
    setUrl(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    filterManager.clear();
    flushSync();

    expect(navigations.map((url) => url.search)).toEqual([""]);
  });

  it("does not undo a url change the seed has not caught up with", () => {
    const { filterManager, navigations, url, seedParams, setUrl } =
      createUrlSyncedFilterManager();
    setUrl(
      perMetricsViewParams({
        [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Google')`,
      }),
    );

    // A back navigation: the page url moves first, the seed follows once
    // `CanvasEntity.onUrlChange` has resolved its redirects.
    const previous = perMetricsViewParams({
      [AD_BIDS_METRICS_NAME]: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook')`,
    });
    url.set(new URL(`http://localhost/canvas?${previous.toString()}`));
    flushSync();
    expect(navigations).toEqual([]);

    seedParams.set(previous);
    flushSync();

    expect(navigations).toEqual([]);
    expect(filterManager.exprByMetricsView[AD_BIDS_METRICS_NAME]).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook"]),
      ]),
    );
  });
});
