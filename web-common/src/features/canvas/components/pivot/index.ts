import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import CanvasPivotDisplay from "./CanvasPivotDisplay.svelte";
import type { TableSpec } from "../table";
import { get } from "svelte/store";

export interface PivotSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measures: string[];
  row_dimensions?: string[];
  col_dimensions?: string[];
}

export { default as Pivot } from "./CanvasPivotDisplay.svelte";

export class PivotCanvasComponent extends BaseCanvasComponent<PivotSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 10 };
  resetParams = ["measures", "row_dimensions", "col_dimensions"];
  type: CanvasComponentType = "pivot";
  component = CanvasPivotDisplay;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: PivotSpec = {
      metrics_view: "",
      measures: [],
      row_dimensions: [],
      col_dimensions: [],
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: PivotSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  inputParams(): InputParams<PivotSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measures: {
          type: "multi_fields",
          meta: { allowedTypes: ["measure"] },
          label: "Measures",
        },
        col_dimensions: {
          type: "multi_fields",
          meta: { allowedTypes: ["time", "dimension"] },
          label: "Column dimensions",
        },
        row_dimensions: {
          type: "multi_fields",
          meta: { allowedTypes: ["time", "dimension"] },
          label: "Row dimensions",
        },
        ...commonOptions,
      },
      filter: getFilterOptions(true, false),
    };
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): PivotSpec {
    const firstDimension = metricsViewSpec?.dimensions?.[0]?.name;
    const secondDimension = metricsViewSpec?.dimensions?.[1]?.name;

    return {
      metrics_view: metricsViewName,
      measures:
        metricsViewSpec?.measures?.slice(0, 3).map((m) => m.name as string) ??
        [],
      row_dimensions: firstDimension ? [firstDimension] : [],
      col_dimensions: secondDimension ? [secondDimension] : undefined,
    };
  }

  updateTableType(
    newTableType: "pivot" | "table",
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ) {
    if (!this.parent.fileArtifact) return;

    const parentPath = this.pathInYAML.slice(0, -1);
    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);
    const { updateEditorContent } = this.parent.fileArtifact;

    const currentSpec = get(this.specStore);

    const allMeasures =
      metricsViewSpec?.measures?.map((m) => m.name as string) || [];
    const allDimensions =
      metricsViewSpec?.dimensions?.map((d) => d.name || (d.column as string)) ||
      [];

    let newSpec: PivotSpec | TableSpec;
    if (newTableType === "pivot") {
      // Switch to pivot table spec
      const flatTableSpec = currentSpec;
      const { columns = [], ...restFlatTableSpec } = flatTableSpec || {};

      const row_dimensions =
        columns?.filter((c) => allDimensions.includes(c)) || [];
      const measures = columns?.filter((c) => allMeasures.includes(c)) || [];

      newSpec = {
        ...restFlatTableSpec,
        row_dimensions,
        measures,
      };
    } else {
      // Switch to flat table spec
      const pivotTableSpec = currentSpec as unknown as PivotSpec;

      const {
        row_dimensions = [],
        col_dimensions = [],
        measures = [],
        ...restPivotTableSpec
      } = pivotTableSpec || {};

      const columns = [...row_dimensions, ...col_dimensions, ...measures];

      newSpec = {
        ...restPivotTableSpec,
        columns,
      };
    }

    const width = parsedDocument.getIn([...parentPath, "width"]);

    parsedDocument.setIn(parentPath, { [newTableType]: newSpec, width });

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), false, true);
  }
}
