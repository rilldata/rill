import { getFiltersFromText } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  mergeDimensionAndMeasureFilters,
  splitWhereFilter,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { toggleDimensionFilterValue } from "@rilldata/web-common/features/dashboards/state-managers/actions/dimension-filters.ts";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
  forEachIdentifier,
  getValuesInExpression,
  isExpressionUnsupported,
  negateExpression,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  MetricsViewSpecDimension,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import {
  type MetricsViewSpecMeasure,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { get, writable, type Writable } from "svelte/store";
import type { DimensionFilterItem } from "../../dashboards/state-managers/selectors/dimension-filters";
import { DimensionFilterMode } from "../../dashboards/filters/dimension-filters/constants";
import type { MeasureFilterItem } from "../../dashboards/state-managers/selectors/measure-filters";
import type { DimensionThresholdFilter } from "../../dashboards/stores/explore-state";
import { convertExpressionToFilterParam } from "../../dashboards/url-state/filters/converters";
import { FilterManager, type UIFilters } from "./filter-manager";
import { getDimensionDisplayName } from "../../dashboards/filters/getDisplayName";

export type ParsedFilters = ReturnType<typeof initFilterBase>;

export function initFilterBase(metricsViewName: string) {
  return {
    where: createAndExpression([]),
    dimensionFilter: createAndExpression([]),
    urlFormat: undefined as string | undefined,
    metricsViewName,
    dimensionsWithInListFilter: <string[]>[],
    dimensionThresholdFilters: <DimensionThresholdFilter[]>[],
    measureFilters: new Map<string, MeasureFilterItem>(),
    dimensionFilters: new Map<string, DimensionFilterItem>(),
    complexFilters: [] as V1Expression[],
    hasFilters: false,
    hasClearableFilters: false,
  };
}

// Exists at the global and widget level on Canvas
export class FilterState {
  parsed: Writable<ParsedFilters>;
  parsedDefaultFilters: Writable<ParsedFilters>;
  temporaryFilterKeys = writable(new Set<string>());

  constructor(
    private metricsViewName: string,
    private manager: FilterManager,
    public instanceId: string,
  ) {
    this.parsed = writable(initFilterBase(this.metricsViewName));
    this.parsedDefaultFilters = writable(initFilterBase(this.metricsViewName));
  }

  setTemporaryFilterName = (key: string) => {
    const keys = get(this.temporaryFilterKeys);
    keys.add(key);
    this.temporaryFilterKeys.set(keys);

    this.reprocessExistingFilters();
  };

  reprocessExistingFilters = () => {
    const parsed = get(this.parsed);

    this.parsed.set(
      this.parseFilter({
        expr: parsed.where,
        dimensionsWithInListFilter: parsed.dimensionsWithInListFilter,
      }),
    );
  };

  onFilterStringChange(filterString: string) {
    const { urlFormat } = get(this.parsed);
    if (urlFormat === filterString) return;

    this.parsed.set(this.parseFilterString(filterString));
  }

  onDefaultExpressionChange = (expr: V1Expression | undefined) => {
    expr = expr ?? createAndExpression([]);

    this.parsedDefaultFilters.set(
      this.parseFilter({
        expr,
      }),
    );
  };

  clearAllFilters = () => {
    this.parsed.set(this.parseFilterString(""));
    this.temporaryFilterKeys.set(new Set());
    return "";
  };

  parseFilter({
    expr,
    filterString,
    dimensionsWithInListFilter,
  }: {
    expr: V1Expression;
    filterString?: string;
    dimensionsWithInListFilter?: string[];
  }): ParsedFilters {
    const where = structuredClone(expr);
    const { dimensionThresholdFilters, dimensionFilters } =
      splitWhereFilter(expr);

    const isComplexFilter = false;

    filterString =
      filterString ||
      getFilterParam(
        expr,
        dimensionThresholdFilters,
        dimensionsWithInListFilter ?? [],
      ) ||
      "";

    dimensionsWithInListFilter =
      dimensionsWithInListFilter ??
      getFiltersFromText(filterString).dimensionsWithInlistFilter;

    if (isComplexFilter) {
      return {
        urlFormat: filterString,
        where: where,
        dimensionFilter: dimensionFilters,
        metricsViewName: this.metricsViewName,
        dimensionsWithInListFilter,
        dimensionThresholdFilters,
        dimensionFilters: new Map(),
        measureFilters: new Map(),
        complexFilters: [expr],
        hasClearableFilters: false,
        hasFilters: false,
      };
    }

    const dimensionMap =
      this.manager.metricsViewNameDimensionMap.get(this.metricsViewName) ??
      new Map<string, MetricsViewSpecDimension>();
    const measureMap =
      this.manager.metricsViewNameMeasureMap.get(this.metricsViewName) ??
      new Map<string, MetricsViewSpecMeasure>();

    const processed = processExpression({
      expr: dimensionFilters,
      measureMap,
      dimensionMap,
      metricsViewName: this.metricsViewName,
      dimensionsWithInListFilter,
      dimensionThresholdFilters,
      temporaryFilterKeys: get(this.temporaryFilterKeys),
    });

    return {
      urlFormat: filterString,
      where: where,
      dimensionFilter: dimensionFilters,
      metricsViewName: this.metricsViewName,
      dimensionsWithInListFilter,
      dimensionThresholdFilters,
      ...processed,
      complexFilters: [],
    };
  }

  parseFilterString = (filterString: string | undefined) => {
    const { expr, dimensionsWithInlistFilter: dimensionsWithInListFilter } =
      getFiltersFromText(filterString ?? "");

    return this.parseFilter({ expr, filterString, dimensionsWithInListFilter });
  };

  removeDimensionFilter = (dimensionName: string) => {
    const {
      dimensionFilter: df,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    } = get(this.parsed);
    const exprIdx = df.cond?.exprs?.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );
    if (!(exprIdx === undefined || exprIdx === -1)) {
      df.cond?.exprs?.splice(exprIdx, 1);
    }

    return getFilterParam(
      df,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    );
  };

  applyDimensionContainsMode = (dimensionName: string, searchText: string) => {
    const {
      dimensionFilter: wf,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    } = get(this.parsed);

    const exprIndex = wf.cond?.exprs?.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );

    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(
        createLikeExpression(dimensionName, `%${searchText}%`, false),
      );
    } else {
      const operation = wf.cond!.exprs![exprIndex].cond!.op;
      const isExclude =
        operation === V1Operation.OPERATION_NLIKE ||
        operation === V1Operation.OPERATION_NIN;
      wf.cond!.exprs![exprIndex] = createLikeExpression(
        dimensionName,
        `%${searchText}%`,
        isExclude,
      );
    }
    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    );
  };

  toggleDimensionFilterMode = (dimensionName: string) => {
    const {
      dimensionsWithInListFilter,
      dimensionFilter: wf,
      dimensionThresholdFilters,
    } = get(this.parsed);

    if (!wf.cond?.exprs) return;
    const exprIdx = wf.cond.exprs.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );
    if (exprIdx === -1) return;
    wf.cond.exprs[exprIdx] = negateExpression(wf.cond.exprs[exprIdx]);

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    );
  };

  toggleDimensionValueSelections = (
    dimensionName: string,
    dimensionValues: string[],
    keepPillVisible?: boolean,
    isExclusiveFilter?: boolean,
    exclude: boolean = false,
  ) => {
    const {
      dimensionFilter: wf,
      dimensionsWithInListFilter,
      dimensionThresholdFilters,
    } = get(this.parsed);

    let exprIndex =
      wf.cond?.exprs?.findIndex(
        (e) => e.cond?.exprs?.[0].ident === dimensionName,
      ) ?? -1;
    let expr = wf.cond?.exprs?.[exprIndex];

    const wasLikeFilter =
      expr?.cond?.op === V1Operation.OPERATION_LIKE ||
      expr?.cond?.op === V1Operation.OPERATION_NLIKE;

    if (!expr?.cond?.exprs || wasLikeFilter) {
      expr = createInExpression(dimensionName, [], exclude);
      wf.cond?.exprs?.push(expr);
      exprIndex = wf.cond!.exprs!.length - 1;
    }

    const wasInListFilter = dimensionsWithInListFilter.includes(dimensionName);
    if (wasInListFilter) {
      dimensionsWithInListFilter.filter((d) => d !== dimensionName);
    }

    dimensionValues.forEach((dimensionValue) => {
      toggleDimensionFilterValue(expr, dimensionValue, !!isExclusiveFilter);
    });

    if (expr?.cond?.exprs?.length === 1) {
      wf.cond?.exprs?.splice(exprIndex, 1);
    }

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    );
  };

  applyDimensionInListMode = (dimensionName: string, values: string[]) => {
    const {
      dimensionFilter: wf,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    } = get(this.parsed);
    const isExclude = false;

    const expr = createInExpression(dimensionName, values, isExclude);

    dimensionsWithInListFilter.push(dimensionName);

    const exprIndex =
      wf.cond?.exprs?.findIndex(
        (e) => e.cond?.exprs?.[0].ident === dimensionName,
      ) ?? -1;
    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(expr);
    } else {
      wf.cond!.exprs![exprIndex] = expr;
    }

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInListFilter,
    );
  };

  setMeasureFilter = (
    dimensionName: string,
    filter: MeasureFilterEntry,
    oldDimension: string,
  ) => {
    const {
      dimensionThresholdFilters: dtfs,
      dimensionsWithInListFilter,
      dimensionFilter,
    } = get(this.parsed);

    const dimIdx = dtfs.findIndex(
      (dtf) => dtf.name === (oldDimension || dimensionName),
    );
    let dimThresholdFilter = dtfs[dimIdx];

    if (!dimThresholdFilter) {
      dimThresholdFilter = { name: dimensionName, filters: [] };
      dtfs.push(dimThresholdFilter);
    } else {
      if (oldDimension && oldDimension !== dimensionName) {
        dimThresholdFilter.name = dimensionName;
      }
      const filters = dimThresholdFilter.filters;
      const exprIdx = filters.findIndex((f) => f.measure === filter.measure);
      if (exprIdx !== -1) {
        filters.splice(exprIdx, 1);
      }
    }

    const exprIdx = dimThresholdFilter.filters.findIndex(
      (f) => f.measure === filter.measure,
    );
    if (exprIdx === -1) {
      dimThresholdFilter.filters.push(filter);
    } else {
      dimThresholdFilter.filters.splice(exprIdx, 1, filter);
    }

    return getFilterParam(dimensionFilter, dtfs, dimensionsWithInListFilter);
  };
  removeMeasureFilter = (dimensionName: string, measureName: string) => {
    const {
      dimensionThresholdFilters: dtfs,
      dimensionsWithInListFilter,
      dimensionFilter,
    } = get(this.parsed);

    const dimIdx = dtfs.findIndex((dtf) => dtf.name === dimensionName);
    const dimThresholdFilter = dtfs[dimIdx];

    if (dimThresholdFilter) {
      const filters = dimThresholdFilter.filters;
      const exprIdx = filters.findIndex((f) => f.measure === measureName);
      if (exprIdx !== -1) {
        filters.splice(exprIdx, 1);
      }
      if (filters.length === 0) {
        dtfs.splice(dimIdx, 1);
      }
    }

    return getFilterParam(dimensionFilter, dtfs, dimensionsWithInListFilter);
  };
}

function processExpression({
  expr,
  dimensionMap,
  measureMap,
  metricsViewName,
  dimensionsWithInListFilter,
  dimensionThresholdFilters,
  temporaryFilterKeys,
}: {
  expr: V1Expression;
  measureMap: Map<string, MetricsViewSpecMeasure>;
  dimensionMap: Map<string, MetricsViewSpecDimension>;
  metricsViewName: string;
  dimensionsWithInListFilter: string[];
  dimensionThresholdFilters: DimensionThresholdFilter[];
  temporaryFilterKeys: Set<string>;
}): UIFilters {
  const isComplex = isExpressionUnsupported(expr);
  const dimensionFilters = getDimensionFilterItemsMap(
    dimensionMap,
    expr,
    dimensionsWithInListFilter,
    metricsViewName,
  );
  const measureFilters = getCanvasMeasureFiltersMap(
    measureMap,
    dimensionThresholdFilters,
  );
  temporaryFilterKeys.forEach((key) => {
    if (dimensionFilters.has(key)) {
      temporaryFilterKeys.delete(key);
      return;
    }
    const dimension = dimensionMap.get(key);
    if (dimension) {
      dimensionFilters.set(key, {
        name: key,
        label: getDimensionDisplayName(dimension),
        dimensions: new Map([[metricsViewName, dimension]]),
        selectedValues: [],
        mode: DimensionFilterMode.Select,
        isInclude: true,
        pinned: false,
      });
    }

    const measure = measureMap.get(key);
    if (measureFilters.has(key)) {
      temporaryFilterKeys.delete(key);
      return;
    }
    if (measure) {
      measureFilters.set(key, {
        dimensionName: "",
        dimensions: undefined,
        name: key,
        label: measure.displayName ?? "",
        pinned: false,
        measures: new Map([[metricsViewName, measure]]),
        metricsViewNames: [metricsViewName],
      });
    }
  });
  return {
    complexFilters: isComplex ? [expr] : [],
    measureFilters: measureFilters,
    dimensionFilters: dimensionFilters,
    hasFilters: dimensionFilters.size > 0 || measureFilters.size > 0,
    hasClearableFilters: dimensionFilters.size > 0 || measureFilters.size > 0,
  };
}

export function getCanvasMeasureFiltersMap(
  measureIdMap: Map<string, MetricsViewSpecMeasure>,
  dimensionThresholdFilters: DimensionThresholdFilter[],
): Map<string, MeasureFilterItem> {
  const map = new Map();

  dimensionThresholdFilters.forEach((dtf) => {
    const filter = dtf.filters[0];
    const measureName = filter.measure;
    const measure = measureIdMap.get(measureName);
    if (!measure) return;

    const entry: MeasureFilterItem = {
      dimensionName: dtf.name,
      name: measureName,
      label: measure.displayName || measure.expression || filter.measure,
      filter: filter,
    };

    map.set(measureName, entry);
  });

  return map;
}

export function getDimensionFilterItemsMap(
  dimensionIdMap: Map<string, MetricsViewSpecDimension>,
  filter: V1Expression | undefined,
  dimensionsWithInlistFilter: string[],
  metricsViewName: string,
): Map<string, DimensionFilterItem> {
  if (!filter) return new Map();
  const filteredDimensions: Map<string, DimensionFilterItem> = new Map();
  const addedDimension = new Set<string>();

  forEachIdentifier(filter, (e, ident) => {
    if (addedDimension.has(ident) || !dimensionIdMap.has(ident)) return;
    const dim = dimensionIdMap.get(ident);

    if (!dim) {
      return;
    }
    addedDimension.add(ident);

    const op = e.cond?.op;
    if (op === V1Operation.OPERATION_IN || op === V1Operation.OPERATION_NIN) {
      const isInListMode = dimensionsWithInlistFilter.includes(ident);
      filteredDimensions.set(ident, {
        name: ident,
        label: getDimensionDisplayName(dim),
        dimensions: new Map([[metricsViewName, dim]]),
        mode: isInListMode
          ? DimensionFilterMode.InList
          : DimensionFilterMode.Select,
        selectedValues: getValuesInExpression(e),
        isInclude: e.cond?.op === V1Operation.OPERATION_IN,
        inputText: undefined,
        pinned: false,
      });
    } else if (
      op === V1Operation.OPERATION_LIKE ||
      op === V1Operation.OPERATION_NLIKE
    ) {
      filteredDimensions.set(ident, {
        name: ident,
        label: getDimensionDisplayName(dim),
        mode: DimensionFilterMode.Contains,
        selectedValues: [],
        inputText: e.cond?.exprs?.[1]?.val?.toString?.() ?? "",
        isInclude: e.cond?.op === V1Operation.OPERATION_LIKE,
        dimensions: new Map([[metricsViewName, dim]]),
        pinned: false,
      });
    }
  });

  return filteredDimensions;
}

export function getFilterParam(
  whereFilter: V1Expression,
  dtf: DimensionThresholdFilter[],
  dimensionsWithInlistFilter: string[],
) {
  const mergedFilters =
    sanitiseExpression(
      mergeDimensionAndMeasureFilters(
        whereFilter ?? createAndExpression([]),
        dtf,
      ),
      undefined,
    ) ?? createAndExpression([]);

  return convertExpressionToFilterParam(
    mergedFilters,
    dimensionsWithInlistFilter,
  );
}
