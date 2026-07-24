import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/dimension-filter-manager.svelte.ts";
import type { V1Expression } from "@rilldata/web-admin/client";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  forEachIdentifier,
  isExpressionUnsupported,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/measure-filter-manager.svelte.ts";

export class ExpressionFilterManager {
  public measureFilterManagers: MeasureFilterManager[] = $state([]);
  public dimensionFilterManagers: DimensionFilterManager[] = $state([]);
  public temporaryFilter:
    | MeasureFilterManager
    | DimensionFilterManager
    | undefined = $state(undefined);

  public expr: V1Expression;
  public exprParam: string = "";
  public isComplexFilter: boolean = $state(false);

  public exploreName: string = "";
  public metricsViewName: string = "";
  public measureIdMap: Map<string, MetricsViewSpecMeasure> = new Map();
  public dimensionIdMap: Map<string, MetricsViewSpecDimension> = new Map();

  public constructor() {
    this.expr = $derived.by(() => {
      const dimExprs = this.dimensionFilterManagers
        .filter((dfm) => !!dfm.expr)
        .map((d) => d.expr as V1Expression);
      const mesExprs = this.measureFilterManagers
        .filter((mfm) => !!mfm.expr)
        .map((m) => m.expr as V1Expression);
      return createAndExpression(dimExprs.concat(mesExprs));
    });
  }

  public syncSpec(
    metricsViewName: string,
    exploreName: string,
    measureIdMap: Map<string, MetricsViewSpecMeasure>,
    dimensionIdMap: Map<string, MetricsViewSpecDimension>,
  ) {
    this.metricsViewName = metricsViewName;
    this.exploreName = exploreName;
    this.measureIdMap = measureIdMap;
    this.dimensionIdMap = dimensionIdMap;
  }

  public setExprParam(exprParam: string) {
    if (exprParam === this.exprParam) return;
    this.exprParam = exprParam;

    const { expr, dimensionsWithInlistFilter } =
      convertFilterParamToExpression(exprParam);
    this.isComplexFilter = expr ? isExpressionUnsupported(expr) : false;
    if (!expr || this.isComplexFilter) {
      this.dimensionFilterManagers = [];
      this.measureFilterManagers = [];
      return;
    }

    const addedMeasure = new Set<string>();
    const newMeasureFilterManagers: MeasureFilterManager[] = [];

    const addedDimension = new Set<string>();
    const newDimensionFilterManagers: DimensionFilterManager[] = [];

    forEachIdentifier(expr, (e, ident) => {
      const dim = this.dimensionIdMap.get(ident);
      if (!dim) return;

      const firstValueExpr = e?.cond?.exprs?.[1];

      if (firstValueExpr?.subquery) {
        const measureName = firstValueExpr.subquery.measures?.[0];
        if (!measureName) return;

        const measure = this.measureIdMap.get(measureName);
        if (!measure) return;

        addedMeasure.add(measureName);

        newMeasureFilterManagers.push(
          new MeasureFilterManager(
            measureName,
            getMeasureDisplayName(measure),
            new Map<string, MetricsViewSpecMeasure>([
              [this.metricsViewName, measure],
            ]),
            false,
            ident,
            firstValueExpr,
          ),
        );
      } else {
        addedDimension.add(ident);

        const isInListMode = dimensionsWithInlistFilter.includes(ident);
        newDimensionFilterManagers.push(
          new DimensionFilterManager(
            ident,
            getDimensionDisplayName(dim),
            new Map<string, MetricsViewSpecDimension>([
              [this.metricsViewName, dim],
            ]),
            false,
            e,
            isInListMode,
          ),
        );
      }
    });

    if (this.temporaryFilter instanceof MeasureFilterManager) {
      if (!addedMeasure.has(this.temporaryFilter.name)) {
        newMeasureFilterManagers.push(this.temporaryFilter);
      } else {
        this.temporaryFilter = undefined;
      }
    } else if (this.temporaryFilter instanceof DimensionFilterManager) {
      if (!addedDimension.has(this.temporaryFilter.name)) {
        newDimensionFilterManagers.push(this.temporaryFilter);
      } else {
        this.temporaryFilter = undefined;
      }
    }

    this.measureFilterManagers = newMeasureFilterManagers;
    this.dimensionFilterManagers = newDimensionFilterManagers;
  }

  public addTemporaryMeasureFilter(name: string) {
    const measure = this.measureIdMap.get(name);
    if (!measure) return;

    this.temporaryFilter = new MeasureFilterManager(
      name,
      getMeasureDisplayName(measure),
      new Map<string, MetricsViewSpecMeasure>([
        [this.metricsViewName, measure],
      ]),
      false,
    );
    this.measureFilterManagers = [
      ...this.measureFilterManagers,
      this.temporaryFilter,
    ];
  }

  public addTemporaryDimensionFilter(name: string) {
    const dim = this.dimensionIdMap.get(name);
    if (!dim) return;

    this.temporaryFilter = new DimensionFilterManager(
      name,
      getDimensionDisplayName(dim),
      new Map<string, MetricsViewSpecDimension>([[this.metricsViewName, dim]]),
      false,
    );
    this.dimensionFilterManagers = [
      ...this.dimensionFilterManagers,
      this.temporaryFilter,
    ];
  }

  public dimensionFilterAction(
    name: string,
    callback: (dimensionFilterManager: DimensionFilterManager) => any,
  ) {
    const dim = this.dimensionIdMap.get(name);
    if (!dim) return;

    const existingManager = this.dimensionFilterManagers.find(
      (dfm) => dfm.name === name,
    );
    const dimensionFilterManager =
      existingManager ??
      new DimensionFilterManager(
        name,
        getDimensionDisplayName(dim),
        new Map<string, MetricsViewSpecDimension>([
          [this.metricsViewName, dim],
        ]),
        false,
      );

    const ret = callback(dimensionFilterManager);
    if (!existingManager && dimensionFilterManager.expr) {
      this.dimensionFilterManagers.push(dimensionFilterManager);
    }
    return ret;
  }

  public clear() {
    this.measureFilterManagers = [];
    this.dimensionFilterManagers = [];
    this.temporaryFilter = undefined;
  }

  public getOtherDimensionsFilter(name: string) {
    const exprs = this.dimensionFilterManagers
      .filter((dfm) => !!dfm.expr && dfm.name !== name)
      .map((d) => d.expr as V1Expression);
    return exprs.length === 0 ? undefined : createAndExpression(exprs);
  }
}
