import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  getCommonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  ColorRangeMapping,
  FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import {
  type V1MetricsViewSpec,
  type V1Resource,
  MetricsViewSpecDimensionType,
} from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import { get } from "svelte/store";
import CanvasMap from "./CanvasMap.svelte";
import {
  guessRegionBoundaries,
  type RegionBoundaries,
} from "./regions/registry";

export { default as CanvasMap } from "./CanvasMap.svelte";

/** Color on a map is always driven by a measure. */
export interface MapColorConfig {
  measure: string;
  colorRange?: ColorRangeMapping;
}

export const DEFAULT_MAP_COLOR_RANGE: ColorRangeMapping = {
  mode: "scheme",
  scheme: "tealblues",
};

/** Regions such as countries or states, filled by a measure and matched through a dimension's values. */
export interface MapRegionsConfig {
  dimension: string;
  boundaries: RegionBoundaries;
  /** Region property the dimension values are matched against. Defaults to matching every key of the boundary set. */
  key?: string;
  color: MapColorConfig;
}

/** Camera position the map opens with. When set, the map never auto-fits to the data. */
export interface MapInitialView {
  longitude: number;
  latitude: number;
  zoom?: number;
}

export interface MapSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  // Points or polygons from a geo dimension. `color` and the fields below it only apply to this layer.
  geo_dimension?: FieldConfig<"nominal">;
  color?: MapColorConfig;
  size_measure?: FieldConfig<"quantitative">;
  tooltip_dimension?: FieldConfig<"nominal">;
  regions?: MapRegionsConfig;
  initial_view?: MapInitialView;
}

export class MapComponent extends BaseCanvasComponent<MapSpec> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 6, height: 4 };
  resetParams = [
    "geo_dimension",
    "color",
    "size_measure",
    "tooltip_dimension",
    "regions",
    "initial_view",
  ];
  type: CanvasComponentType = "map";
  component = CanvasMap;
  _isPolygonMode = false;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: MapSpec = {
      metrics_view: "",
      geo_dimension: { field: "", type: "nominal" },
      color: { measure: "", colorRange: DEFAULT_MAP_COLOR_RANGE },
    };
    super(resource, parent, path, defaultSpec);
  }

  // A map needs at least one complete layer.
  isValid(spec: MapSpec): boolean {
    if (typeof spec.metrics_view !== "string") return false;
    const hasPoints = !!spec.geo_dimension?.field && !!spec.color?.measure;
    const hasRegions =
      !!spec.regions?.dimension &&
      !!spec.regions?.boundaries &&
      !!spec.regions?.color?.measure;
    return hasPoints || hasRegions;
  }

  inputParams(): InputParams<MapSpec> {
    const hasGeoDimension = !!get(this.specStore)?.geo_dimension?.field;
    const inputParams: InputParams<MapSpec> = {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        regions: {
          type: "map_regions",
          optional: true,
          label: "Regions",
        },
        geo_dimension: {
          type: "positional",
          optional: true,
          label: "Geo dimension",
          meta: {
            chartFieldInput: {
              type: "dimension",
              geoOnly: true,
              hideTimeDimension: true,
              isRemovable: true,
            },
          },
        },
        color: {
          type: "map_color",
          label: "Color",
          showInUI: hasGeoDimension,
        },
        size_measure: {
          type: "positional",
          optional: true,
          label: "Size measure",
          showInUI: hasGeoDimension && !this._isPolygonMode,
          meta: {
            chartFieldInput: {
              type: "measure",
              isRemovable: true,
            },
          },
        },
        tooltip_dimension: {
          type: "positional",
          optional: true,
          label: "Tooltip dimension",
          showInUI: hasGeoDimension,
          meta: {
            chartFieldInput: {
              type: "dimension",
              hideTimeDimension: true,
              isRemovable: true,
            },
          },
        },
        ...getCommonOptions(),
      },
      filter: getFilterOptions(),
    };

    return inputParams;
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): MapSpec {
    const geoDimension = metricsViewSpec?.dimensions?.find(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_GEOSPATIAL,
    );
    const geoDimensionName = geoDimension?.name || "";

    // Color is required, so fall back to the first measure in the metrics view
    const colorMeasure = metricsViewSpec?.measures?.[0]?.name ?? "";

    // Without a geo dimension, look for one that names regions, such as `country`.
    if (!geoDimensionName) {
      for (const d of metricsViewSpec?.dimensions ?? []) {
        const boundaries = d.name ? guessRegionBoundaries(d.name) : undefined;
        if (!d.name || !boundaries) continue;
        return {
          metrics_view: metricsViewName,
          regions: {
            dimension: d.name,
            boundaries,
            color: {
              measure: colorMeasure,
              colorRange: DEFAULT_MAP_COLOR_RANGE,
            },
          },
        };
      }
    }

    return {
      metrics_view: metricsViewName,
      geo_dimension: { field: geoDimensionName, type: "nominal" },
      color: {
        measure: colorMeasure,
        colorRange: DEFAULT_MAP_COLOR_RANGE,
      },
    };
  }

  get isBuilderMode(): boolean {
    return !!this.parent.fileArtifact;
  }
}
