import type { ChartType } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { PivotSpec } from "@rilldata/web-common/features/canvas/components/pivot";
import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
import type { ComponentSize } from "@rilldata/web-common/features/canvas/components/types";
import { getParsedDocument } from "@rilldata/web-common/features/canvas/inspector/selectors";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { get, writable, type Writable } from "svelte/store";

// A base class that implements all the store logic
export abstract class BaseCanvasComponent<T> {
  /**
   * Local copy of the spec as a svelte writable store
   */
  specStore: Writable<T>;
  /**
   * Path in the YAML where the component is stored
   */
  pathInYAML: (string | number)[] = [];
  /**
   * File artifact where the component
   * is stored
   */
  fileArtifact: FileArtifact | undefined = undefined;

  // Let child classes define these
  /**
   * Minimum allowed size for the component
   * container on the canvas
   */
  abstract minSize: ComponentSize;

  /**
   * The default size of the container when the component
   * is added to the canvas
   */
  abstract defaultSize: ComponentSize;

  /**
   * The parameters that should be reset when the metrics_view
   * is changed
   */
  abstract resetParams: string[];

  /**
   * The minimum condition needed for the spec to be valid
   * for the given component and to be rendered on the canvas
   */
  abstract isValid(spec: T): boolean;

  /**
   * A map of input params which will be used in the visual
   * UI builder
   */
  abstract inputParams(): InputParams<T>;

  /**
   * Get the spec when the component is added to the canvas
   */
  abstract newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): T;

  constructor(
    fileArtifact: FileArtifact | undefined,
    path: (string | number)[],
    defaultSpec: T,
    initialSpec: Partial<T> = {},
  ) {
    // Initialize the store with merged spec
    const mergedSpec = { ...defaultSpec, ...initialSpec };
    this.specStore = writable(mergedSpec);
    this.pathInYAML = path;
    this.fileArtifact = fileArtifact;
  }

  private async updateYAML(newSpec: T): Promise<void> {
    if (!this.fileArtifact) return;
    const parseDocumentStore = getParsedDocument(this.fileArtifact);
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent, saveLocalContent } = this.fileArtifact;

    // Update the Item
    parsedDocument.setIn(this.pathInYAML, newSpec);

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), false);
    await saveLocalContent();
  }

  /**
   * Set the spec store and YAML with the new values
   */
  async setSpec(newSpec: T): Promise<void> {
    if (this.isValid(newSpec)) {
      await this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }

  /**
   * Update the spec store and YAML with the new values
   */
  // TODO: Add stricter type definition for keys and value deriving from spec
  async updateProperty(key: string, value: unknown): Promise<void> {
    const currentSpec = get(this.specStore);

    const newSpec = { ...currentSpec, [key]: value };

    if (value === undefined || value == "") {
      delete newSpec[key];
    }

    // If the metrics_view is changed, clear the time_filters and dimension_filters
    if (key === "metrics_view") {
      if ("time_filters" in newSpec) {
        delete newSpec.time_filters;
      }
      if ("dimension_filters" in newSpec) {
        delete newSpec.dimension_filters;
      }
      if (this.resetParams.length > 0) {
        this.resetParams.forEach((param) => {
          delete newSpec[param];
        });
      }
    }

    if (this.isValid(newSpec)) {
      await this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }

  /**
   * Update the chart type of chart component in store and YAML
   */
  async updateChartType(key: ChartType) {
    if (!this.fileArtifact) return;
    const currentSpec = get(this.specStore);

    const parentPath = this.pathInYAML.slice(0, -1);

    const parseDocumentStore = getParsedDocument(this.fileArtifact);
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent, saveLocalContent } = this.fileArtifact;

    const width = parsedDocument.getIn([...parentPath, "width"]);

    parsedDocument.setIn(parentPath, { [key]: currentSpec, width });

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }

  async updateTableType(
    newTableType: "pivot" | "table",
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ) {
    if (!this.fileArtifact) return;
    const parentPath = this.pathInYAML.slice(0, -1);
    const parseDocumentStore = getParsedDocument(this.fileArtifact);
    const parsedDocument = get(parseDocumentStore);
    const { updateEditorContent, saveLocalContent } = this.fileArtifact;

    const currentSpec = get(this.specStore);

    const allMeasures =
      metricsViewSpec?.measures?.map((m) => m.name as string) || [];
    const allDimensions =
      metricsViewSpec?.dimensions?.map((d) => d.name || (d.column as string)) ||
      [];

    let newSpec: PivotSpec | TableSpec;
    if (newTableType === "pivot") {
      // Switch to pivot table spec
      const flatTableSpec = currentSpec as TableSpec;

      const row_dimensions =
        flatTableSpec?.columns?.filter((c) => allDimensions.includes(c)) || [];
      const measures =
        flatTableSpec?.columns?.filter((c) => allMeasures.includes(c)) || [];

      newSpec = {
        metrics_view: flatTableSpec?.metrics_view,
        row_dimensions,
        measures,
      };
    } else {
      // Switch to flat table spec

      const pivotTableSpec = currentSpec as PivotSpec;

      const row_dimensions = pivotTableSpec.row_dimensions || [];
      const col_dimensions = pivotTableSpec.col_dimensions || [];
      const measures = pivotTableSpec.measures || [];

      newSpec = {
        metrics_view: pivotTableSpec?.metrics_view,
        columns: [...row_dimensions, ...col_dimensions, ...measures],
      };
    }

    const width = parsedDocument.getIn([...parentPath, "width"]);

    parsedDocument.setIn(parentPath, { [newTableType]: newSpec, width });

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }
}
