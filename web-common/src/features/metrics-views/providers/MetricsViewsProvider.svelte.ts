import {
  createQueryServiceMetricsViewTimeRange,
  createRuntimeServiceListResources,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
  type V1Resource,
  V1TimeGrain,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { isSimpleMeasure } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
import { Duration } from "luxon";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { arrayUnorderedEquals } from "@rilldata/web-common/lib/arrayUtils.ts";
import { V1TimeGrainToOrder } from "@rilldata/web-common/lib/time/new-grains.ts";

export type MetricsViewName = string;
export type DimensionName = string;
export type MeasureName = string;

/**
 * Reactive view over a set of metrics views.
 *
 * Specs for every metrics view come from a single ListResources subscription.
 * Time range summaries are fetched per metrics view, and only for the ones that have a time dimension,
 * so the summaries arrive after the specs rather than alongside them.
 *
 * Measures and dimensions are exposed two ways:
 * as deduped flat lists for pickers, and as name -> metrics view -> spec maps for callers that need to
 * know which metrics views a given measure or dimension belongs to.
 */
export class MetricsViewsProvider {
  /** Valid spec per metrics view name. Absent while the resource is loading or invalid. */
  public specs = $state<Record<MetricsViewName, V1MetricsViewSpec>>({});
  /** Time range summary per metrics view name. Absent for metrics views without a time dimension. */
  public timeRangeSummaries = $state<
    Record<MetricsViewName, V1TimeRangeSummary>
  >({});
  /** Max queryable time range in milliseconds per metrics view name. Zero when unrestricted. */
  public maxQueryTimeRangeMillis = $state<Record<string, number>>({});

  /** Dimension spec per metrics view, keyed by dimension name (or column when unnamed). */
  public dimensionSpecs = $state<
    Record<DimensionName, Record<MetricsViewName, MetricsViewSpecDimension>>
  >({});
  /**
   * Measure spec per metrics view, keyed by measure name.
   * The same measure name can be defined by more than one metrics view.
   */
  public measureSpecs = $state<
    Record<MeasureName, Record<MetricsViewName, MetricsViewSpecMeasure>>
  >({});

  /** Deduped by name across metrics views; the first metrics view to define a name wins. */
  public measures = $state<MetricsViewSpecMeasure[]>([]);
  public simpleMeasures = $state<MetricsViewSpecMeasure[]>([]);
  public dimensions = $state<MetricsViewSpecDimension[]>([]);

  /** Union of the individual summaries: earliest min, latest max, latest watermark. */
  public timeRangeSummary: V1TimeRangeSummary | undefined;
  /** Smallest restriction across the metrics views, since it has to hold for all of them. */
  public maxQueryTimeRange: Duration | undefined;
  /** Smallest time grain across the metrics views, since it has to hold for all of them. */
  public smallestTimeGrain: V1TimeGrain | undefined;
  public smallestGrainOrder: number;
  /** True once every metrics view has a spec and every time series metrics view has a summary. */
  public ready: boolean;
  public metricsViewNames = $state<string[]>([]);

  public cleanup: () => void;

  private resources: V1Resource[] = [];
  private readonly timeRangeUnsubs = new Map<string, () => void>();

  public constructor(
    public readonly runtimeClient: RuntimeClient,
    initMetricsViewNames: string[],
  ) {
    this.metricsViewNames = initMetricsViewNames.filter(Boolean);

    const allResourcesQuery = createRuntimeServiceListResources(
      runtimeClient,
      {},
      undefined,
      queryClient,
    );
    const allResourcesUnsub = allResourcesQuery.subscribe(
      (allResourcesResp) => {
        this.resources = allResourcesResp.data?.resources ?? [];
        this.processResources();
      },
    );

    this.timeRangeSummary = $derived.by(() => {
      let min: string | undefined;
      let max: string | undefined;
      let watermark: string | undefined;
      let minTime = Infinity;
      let maxTime = -Infinity;
      let watermarkTime = -Infinity;

      for (const metricsViewName of this.metricsViewNames) {
        const summary = this.timeRangeSummaries[metricsViewName];
        if (!summary) continue;

        // Date.parse returns NaN for missing or malformed timestamps,
        // and every comparison against NaN is false, so those simply never win.
        const minCandidate = Date.parse(summary.min ?? "");
        if (minCandidate < minTime) {
          minTime = minCandidate;
          min = summary.min;
        }

        const maxCandidate = Date.parse(summary.max ?? "");
        if (maxCandidate > maxTime) {
          maxTime = maxCandidate;
          max = summary.max;
        }

        const watermarkCandidate = Date.parse(summary.watermark ?? "");
        if (watermarkCandidate > watermarkTime) {
          watermarkTime = watermarkCandidate;
          watermark = summary.watermark;
        }
      }

      if (!min && !max && !watermark) return undefined;
      return { min, max, watermark };
    });

    this.maxQueryTimeRange = $derived.by(() => {
      let smallestMillis = Infinity;
      for (const metricsViewName of this.metricsViewNames) {
        const millis = this.maxQueryTimeRangeMillis[metricsViewName] ?? 0;
        if (millis > 0 && millis < smallestMillis) smallestMillis = millis;
      }
      return smallestMillis === Infinity
        ? undefined
        : Duration.fromMillis(smallestMillis);
    });

    this.ready = $derived(
      this.metricsViewNames.length > 0 &&
        this.metricsViewNames.every((metricsViewName) => {
          const spec = this.specs[metricsViewName];
          if (!spec) return false;
          return (
            !spec.timeDimension || !!this.timeRangeSummaries[metricsViewName]
          );
        }),
    );

    this.cleanup = () => {
      allResourcesUnsub();
      this.timeRangeUnsubs.forEach((unsub) => unsub());
      this.timeRangeUnsubs.clear();
    };
  }

  public setMetricsViewNames(metricsViewNames: string[]) {
    metricsViewNames = metricsViewNames.filter(Boolean);
    if (arrayUnorderedEquals(this.metricsViewNames, metricsViewNames)) return;
    this.metricsViewNames = metricsViewNames;
    this.processResources();
  }

  private processResources() {
    const specs: Record<string, V1MetricsViewSpec> = {};

    const measureSpecs: Record<
      string,
      Record<string, MetricsViewSpecMeasure>
    > = {};
    const measures: MetricsViewSpecMeasure[] = [];
    const simpleMeasures: MetricsViewSpecMeasure[] = [];

    const dimensionSpecs: Record<
      string,
      Record<string, MetricsViewSpecDimension>
    > = {};
    const dimensions: MetricsViewSpecDimension[] = [];

    let smallestTimeGrain: V1TimeGrain | undefined = undefined;
    let smallestGrainOrder: number | undefined = Infinity;

    for (const metricsViewName of this.metricsViewNames) {
      const res = this.resources.find(
        (resource) =>
          resource.meta?.name?.name === metricsViewName &&
          resource.meta?.name?.kind === ResourceKind.MetricsView,
      );
      const spec = res?.metricsView?.state?.validSpec;
      if (!spec) continue;
      specs[metricsViewName] = spec;

      spec.measures?.forEach((measure) => {
        if (!measure.name) return;

        let specsForMeasure = measureSpecs[measure.name];
        if (!specsForMeasure) {
          specsForMeasure = measureSpecs[measure.name] = {};
          measures.push(measure);
          if (isSimpleMeasure(measure)) simpleMeasures.push(measure);
        }
        specsForMeasure[metricsViewName] = measure;
      });

      spec.dimensions?.forEach((dimension) => {
        // Filter expressions identify an unnamed dimension by its column.
        const dimensionName = dimension.name || dimension.column;
        if (!dimensionName) return;

        let specsForDimension = dimensionSpecs[dimensionName];
        if (!specsForDimension) {
          specsForDimension = dimensionSpecs[dimensionName] = {};
          dimensions.push(dimension);
        }
        specsForDimension[metricsViewName] = dimension;
      });

      if (spec.smallestTimeGrain) {
        const specGrainOrder = V1TimeGrainToOrder[spec.smallestTimeGrain];

        if (!smallestTimeGrain) {
          smallestTimeGrain = spec.smallestTimeGrain;
          smallestGrainOrder = specGrainOrder;
        } else if (specGrainOrder < smallestGrainOrder) {
          smallestTimeGrain = spec.smallestTimeGrain;
          smallestGrainOrder = specGrainOrder;
        }
      }

      this.subscribeToTimeRange(metricsViewName, spec);
    }

    this.specs = specs;
    this.measureSpecs = measureSpecs;
    this.measures = measures;
    this.simpleMeasures = simpleMeasures;
    this.dimensionSpecs = dimensionSpecs;
    this.dimensions = dimensions;
    this.smallestTimeGrain = smallestTimeGrain;
    this.smallestGrainOrder = smallestTimeGrain
      ? smallestGrainOrder
      : V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MINUTE];
  }

  /**
   * Starts the time range query for a metrics view the first time its spec shows up.
   * Metrics views without a time dimension have no summary to fetch.
   */
  private subscribeToTimeRange(
    metricsViewName: string,
    spec: V1MetricsViewSpec,
  ) {
    if (!spec.timeDimension || this.timeRangeUnsubs.has(metricsViewName)) {
      return;
    }

    const timeRangeQuery = createQueryServiceMetricsViewTimeRange(
      this.runtimeClient,
      { metricsViewName },
      undefined,
      queryClient,
    );
    this.timeRangeUnsubs.set(
      metricsViewName,
      timeRangeQuery.subscribe((timeRangeResp) => {
        const summary = timeRangeResp.data?.timeRangeSummary;
        if (summary) this.timeRangeSummaries[metricsViewName] = summary;
        this.maxQueryTimeRangeMillis[metricsViewName] = Number(
          timeRangeResp.data?.maxQueryTimeRangeMillis ?? 0,
        );
      }),
    );
  }
}
