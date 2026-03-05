import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { ColorRangeMapping } from "@rilldata/web-common/features/components/charts/types";
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

export interface MapColorConfig {
  measure: string;
  colorRange?: ColorRangeMapping;
}

export interface MapSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  geo_dimension: string;
  color: string | MapColorConfig;
  size_measure?: string;
  tooltip_dimension?: string;
}

export function isMapColorConfig(
  color: string | MapColorConfig | undefined,
): color is MapColorConfig {
  return typeof color === "object" && color !== null && "measure" in color;
}

export class MapComponent extends BaseCanvasComponent<MapSpec> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 6, height: 4 };
  resetParams = ["geo_dimension", "color", "size_measure", "tooltip_dimension"];
  type: CanvasComponentType = "map";
  component = CanvasMap;
  _isPolygonMode = false;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: MapSpec = {
      metrics_view: "",
      geo_dimension: "",
      color: "primary",
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: MapSpec): boolean {
    return (
      typeof spec.metrics_view === "string" &&
      typeof spec.geo_dimension === "string" &&
      spec.geo_dimension !== ""
    );
  }

  inputParams(): InputParams<MapSpec> {
    const inputParams: InputParams<MapSpec> = {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        geo_dimension: {
          type: "dimension",
          label: "Geo dimension",
          meta: { geoOnly: true },
        },
        color: {
          type: "map_color",
          label: "Color",
        },
        size_measure: {
          type: "measure",
          optional: true,
          label: "Size measure",
          showInUI: !this._isPolygonMode,
        },
        tooltip_dimension: {
          type: "dimension",
          optional: true,
          label: "Tooltip dimension",
        },
        ...commonOptions,
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

    // Get first measure for color if available
    const firstMeasure = metricsViewSpec?.measures?.[0];
    const colorMeasure = firstMeasure?.name;

    return {
      metrics_view: metricsViewName,
      geo_dimension: geoDimensionName,
      color: colorMeasure
        ? {
            measure: colorMeasure,
            colorRange: { mode: "scheme", scheme: "tealblues" },
          }
        : "primary",
    };
  }

  get isBuilderMode(): boolean {
    return !!this.parent.fileArtifact;
  }
}
