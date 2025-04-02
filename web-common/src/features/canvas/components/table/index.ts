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
import CanvasTableDisplay from "./CanvasTableDisplay.svelte";

import { get } from "svelte/store";
import type { PivotSpec } from "../pivot";

export interface TableSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  columns: string[];
}

export { default as Table } from "./CanvasTableDisplay.svelte";

export class TableCanvasComponent extends BaseCanvasComponent<TableSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 10 };
  resetParams = ["measures", "row_dimensions", "col_dimensions"];
  type: CanvasComponentType = "table";
  component = CanvasTableDisplay;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: TableSpec = {
      metrics_view: "",
      columns: [],
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: TableSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  inputParams(): InputParams<TableSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        columns: {
          type: "multi_fields",
          label: "Columns",
          meta: { allowedTypes: ["time", "dimension", "measure"] },
        },
        ...commonOptions,
      },
      filter: getFilterOptions(true, false),
    };
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): TableSpec {
    const measures =
      metricsViewSpec?.measures?.slice(0, 3).map((m) => m.name as string) ?? [];

    const dimensions =
      metricsViewSpec?.dimensions
        ?.slice(0, 3)
        .map((d) => d.name || (d.column as string)) ?? [];

    return {
      metrics_view: metricsViewName,
      columns: [...dimensions, ...measures],
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
