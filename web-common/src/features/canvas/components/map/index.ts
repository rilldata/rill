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
import CanvasMap from "./CanvasMap.svelte";

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
  geo_dimension: FieldConfig<"nominal">;
  color: MapColorConfig;
  size_measure?: FieldConfig<"quantitative">;
  tooltip_dimension?: FieldConfig<"nominal">;
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

  isValid(spec: MapSpec): boolean {
    return (
      typeof spec.metrics_view === "string" &&
      !!spec.geo_dimension?.field &&
      !!spec.color?.measure
    );
  }

  inputParams(): InputParams<MapSpec> {
    const inputParams: InputParams<MapSpec> = {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        geo_dimension: {
          type: "positional",
          label: "Geo dimension",
          meta: {
            chartFieldInput: {
              type: "dimension",
              geoOnly: true,
              hideTimeDimension: true,
            },
          },
        },
        color: {
          type: "map_color",
          label: "Color",
        },
        size_measure: {
          type: "positional",
          optional: true,
          label: "Size measure",
          showInUI: !this._isPolygonMode,
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
    // Find first geo dimension
    const geoDimension = metricsViewSpec?.dimensions?.find(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_GEOSPATIAL,
    );
    const geoDimensionName = geoDimension?.name || "";

    // Color is required, so fall back to the first measure in the metrics view
    const colorMeasure = metricsViewSpec?.measures?.[0]?.name ?? "";

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
