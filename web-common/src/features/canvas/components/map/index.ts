import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import MapDisplay from "./Map.svelte";

export interface MapSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  dimension: string;
  measure: string;
  geometry_type?: "h3" | "geojson"; // Type of geometry data, auto-detected if not set
  resolution?: number; // H3 resolution (0-15), if set will aggregate to this resolution
  size_measure?: string; // Optional second measure for point size (only for GeoJSON)
  label_dimension?: string; // Optional dimension to display in tooltip
}

export { default as Map } from "./Map.svelte";

export class MapCanvasComponent extends BaseCanvasComponent<MapSpec> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 8, height: 6 };
  resetParams = ["dimension", "measure"];
  type: CanvasComponentType = "map";
  component = MapDisplay;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultMapSpec: MapSpec = {
      metrics_view: "",
      dimension: "",
      measure: "",
    };

    super(resource, parent, path, defaultMapSpec);
  }

  isValid(spec: MapSpec): boolean {
    return (
      typeof spec.metrics_view === "string" &&
      typeof spec.dimension === "string" &&
      typeof spec.measure === "string" &&
      spec.metrics_view.length > 0 &&
      spec.dimension.length > 0 &&
      spec.measure.length > 0
    );
  }

  getExploreTransformerProperties(): Partial<ExploreState> {
    return {
      activePage: DashboardState_ActivePage.PIVOT,
    };
  }

  inputParams(): InputParams<MapSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        dimension: {
          type: "dimension",
          label: "Dimension (Geometry)",
        },
        measure: {
          type: "measure",
          label: "Measure (for color)",
        },
        geometry_type: {
          type: "switcher_tab",
          label: "Geometry Type",
          optional: true,
          meta: {
            options: [
              { value: "h3", label: "H3" },
              { value: "geojson", label: "GeoJSON" },
            ],
            default: "h3",
          },
        },
        resolution: {
          type: "number",
          label: "H3 Resolution",
          optional: true,
          meta: {
            placeholder: "Auto (from data)",
            showIf: (spec: MapSpec) => spec.geometry_type === "h3" || !spec.geometry_type,
          },
        },
        size_measure: {
          type: "measure",
          label: "Size Measure (optional)",
          optional: true,
          meta: {
            placeholder: "Select for point sizing",
            showIf: (spec: MapSpec) => spec.geometry_type === "geojson",
          },
        },
        label_dimension: {
          type: "dimension",
          label: "Label Dimension (optional)",
          optional: true,
          meta: {
            placeholder: "Show in tooltip",
          },
        },
        ...commonOptions,
      },
      filter: getFilterOptions(true, false),
    };
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): MapSpec {
    const measures = metricsViewSpec?.measures || [];
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure =
      (measures[Math.floor(Math.random() * measures.length)]?.name as string) ||
      "";
    const randomDimension =
      (dimensions[Math.floor(Math.random() * dimensions.length)]
        ?.name as string) || "";

    return {
      metrics_view: metricsViewName,
      dimension: randomDimension,
      measure: randomMeasure,
    };
  }
}

