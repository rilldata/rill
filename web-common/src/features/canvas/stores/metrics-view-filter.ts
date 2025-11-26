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
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import {
  type MetricsViewSpecMeasure,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import type { DimensionFilterItem } from "../../dashboards/state-managers/selectors/dimension-filters";
import { DimensionFilterMode } from "../../dashboards/filters/dimension-filters/constants";
import type { MeasureFilterItem } from "../../dashboards/state-managers/selectors/measure-filters";
import type { DimensionThresholdFilter } from "../../dashboards/stores/explore-state";
import { convertExpressionToFilterParam } from "../../dashboards/url-state/filters/converters";
import {
  FilterManager,
  initFilterBase,
  type ParsedFilters,
  type UIFilters,
} from "./filter-manager";

// wip - bgh
export class MetricsViewFilter {
  parsed = writable(initFilterBase());
  parsedDefaultFilters = writable<ParsedFilters>(initFilterBase());

  constructor(
    metricsView: V1MetricsView,
    private metricsViewName: string,
    defaultExpression: string | undefined,
    private manager: FilterManager,
  ) {
    this.update(metricsView, defaultExpression);
  }

  update(metricsView: V1MetricsView, defaultExpression?: string) {
    this.parsedDefaultFilters.set(this.parseFilterString(defaultExpression));
  }

  onFilterStringChange(filterString: string) {
    this.parsed.set(this.parseFilterString(filterString));
  }

  parseFilterString(filterString: string = ""): ParsedFilters {
    const { expr, dimensionsWithInlistFilter } =
      getFiltersFromText(filterString);

    const { dimensionThresholdFilters, dimensionFilters } =
      splitWhereFilter(expr);

    const isComplexFilter = isExpressionUnsupported(expr);

    if (isComplexFilter) {
      return {
        string: filterString,
        where: expr,
        dimensionFilter: dimensionFilters,
        metricsViewName: this.metricsViewName,
        dimensionsWithInlistFilter,
        dimensionThresholdFilters,
        dimensionFilters: new Map(),
        measureFilters: new Map(),
        complex: true,
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
      dimensionsWithInlistFilter,
      dimensionThresholdFilters,
    });

    return {
      string: filterString,
      where: expr,
      dimensionFilter: dimensionFilters,
      metricsViewName: this.metricsViewName,
      dimensionsWithInlistFilter,
      dimensionThresholdFilters,
      ...processed,
      complex: false,
    };
  }

  removeDimensionFilter = (dimensionName: string) => {
    const {
      where: wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    } = get(this.parsed);
    const exprIdx = wf.cond?.exprs?.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );
    if (!(exprIdx === undefined || exprIdx === -1)) {
      wf.cond?.exprs?.splice(exprIdx, 1);
    }

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    );
  };

  applyDimensionContainsMode = (dimensionName: string, searchText: string) => {
    const {
      where: wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
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
      dimensionsWithInlistFilter,
    );
  };

  toggleDimensionFilterMode = (dimensionName: string) => {
    const {
      dimensionsWithInlistFilter,
      where: wf,
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
      dimensionsWithInlistFilter,
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
      where: wf,
      dimensionsWithInlistFilter,
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

    const wasInListFilter = dimensionsWithInlistFilter.includes(dimensionName);
    if (wasInListFilter) {
      dimensionsWithInlistFilter.filter((d) => d !== dimensionName);
    }

    dimensionValues.forEach((dimensionValue) => {
      toggleDimensionFilterValue(expr, dimensionValue, !!isExclusiveFilter);
    });

    if (expr?.cond?.exprs?.length === 1) {
      wf.cond?.exprs?.splice(exprIndex, 1);

      if (keepPillVisible) {
        // this.setTemporaryFilterName(dimensionName);
      }
    }

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    );
  };

  applyDimensionInListMode = (dimensionName: string, values: string[]) => {
    const {
      where: wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    } = get(this.parsed);
    const isExclude = false;

    const expr = createInExpression(dimensionName, values, isExclude);

    dimensionsWithInlistFilter.push(dimensionName);

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
      dimensionsWithInlistFilter,
    );
  };

  setMeasureFilter = (dimensionName: string, filter: MeasureFilterEntry) => {
    const {
      dimensionThresholdFilters: dtfs,
      dimensionsWithInlistFilter,
      dimensionFilter,
    } = get(this.parsed);

    const dimIdx = dtfs.findIndex((dtf) => dtf.name === dimensionName);
    let dimThresholdFilter = dtfs[dimIdx];

    if (!dimThresholdFilter) {
      dimThresholdFilter = { name: dimensionName, filters: [] };
      dtfs.push(dimThresholdFilter);
    } else {
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

    return getFilterParam(dimensionFilter, dtfs, dimensionsWithInlistFilter);
  };
  removeMeasureFilter = (dimensionName: string, measureName: string) => {
    const {
      dimensionThresholdFilters: dtfs,
      dimensionsWithInlistFilter,
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

    return getFilterParam(dimensionFilter, dtfs, dimensionsWithInlistFilter);
  };
}

function processExpression({
  expr,
  dimensionMap,
  measureMap,
  metricsViewName,
  dimensionsWithInlistFilter,
  dimensionThresholdFilters,
}: {
  expr: V1Expression;
  measureMap: Map<string, MetricsViewSpecMeasure>;
  dimensionMap: Map<string, MetricsViewSpecDimension>;
  metricsViewName: string;
  dimensionsWithInlistFilter: string[];
  dimensionThresholdFilters: DimensionThresholdFilter[];
}): UIFilters {
  const isComplex = isExpressionUnsupported(expr);
  const dimensionFilters = getDimensionFilterItemsMap(
    dimensionMap,
    expr,
    dimensionsWithInlistFilter,
    metricsViewName,
  );
  const measureFilters = getCanvasMeasureFiltersMap(
    measureMap,
    dimensionThresholdFilters,
  );
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

      // dimensions,
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
