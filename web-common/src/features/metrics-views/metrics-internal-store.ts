import { CATEGORICALS } from "@rilldata/web-common/lib/duckdb-data-types";
import { guidGenerator } from "@rilldata/web-common/lib/guid";
import type {
  V1Model,
  V1ReconcileError,
} from "@rilldata/web-common/runtime-client";
import { readable, Subscriber } from "svelte/store";
import { Document, ParsedNode, parseDocument, YAMLMap } from "yaml";
import type { Collection } from "yaml/dist/nodes/Collection";
import { selectTimestampColumnFromSchema } from "./column-selectors";

export interface MetricsConfig extends MetricsParams {
  measures: MeasureEntity[];
  dimensions: DimensionEntity[];
}
export interface MetricsParams {
  display_name: string;
  timeseries: string;
  smallest_time_grain?: string;
  default_time_range?: string;
  model: string;
}
export interface MeasureEntity {
  label?: string;
  expression?: string;
  description?: string;
  format_preset?: string;
  __GUID__?: string;
  __ERROR__?: string;
}
export interface DimensionEntity {
  label?: string;
  property?: string;
  description?: string;
  __ERROR__?: string;
}

export class MetricsInternalRepresentation {
  // All operations are done on the document to preserve comments
  internalRepresentationDocument: Document.Parsed<ParsedNode>;

  // Object respresentation of the Internal YAML Document
  internalRepresentation: MetricsConfig;

  // String representation of Internal YAML document
  internalYAML: string;

  updateStore: (instance: MetricsInternalRepresentation) => void;

  updateRuntime: (yamlString: string) => void;

  constructor(yamlString: string, updateRuntime) {
    this.internalRepresentation = this.decorateInternalRepresentation(
      yamlString
    ) as MetricsConfig;
    this.internalYAML = yamlString;

    this.updateRuntime = updateRuntime;
  }

  bindStore(updateStore: Subscriber<MetricsInternalRepresentation>) {
    this.updateStore = updateStore;
  }

  decorateInternalRepresentation(yamlString: string) {
    const internalRepresentationDoc = parseDocument(yamlString);
    const numberOfMeasures =
      (internalRepresentationDoc.get("measures") as Collection)?.items
        ?.length || 0;

    Array(numberOfMeasures)
      .fill(0)
      .map((_, i) => {
        const measure = internalRepresentationDoc.getIn([
          "measures",
          i,
        ]) as YAMLMap;

        measure.add({ key: "__GUID__", value: guidGenerator() });
      });

    this.internalRepresentationDocument = internalRepresentationDoc;

    return internalRepresentationDoc.toJSON();
  }

  regenerateInternalYAML(shouldUpdateRuntime = true) {
    // create json before any fields are removed
    this.internalRepresentation = this.internalRepresentationDocument.toJSON();

    // remove fields that are not to be sent as yaml
    const temporaryRepresentation = this.internalRepresentationDocument.clone();

    const numberOfMeasures =
      (temporaryRepresentation.get("measures") as Collection)?.items?.length ||
      0;

    // if no measures, this block is skipped.
    for (let i = 0; i < numberOfMeasures; i++) {
      const measure = temporaryRepresentation.getIn(["measures", i]) as YAMLMap;

      if (measure.has("__GUID__"))
        temporaryRepresentation.deleteIn(["measures", i, "__GUID__"]);

      if (measure.has("__ERROR__"))
        temporaryRepresentation.deleteIn(["measures", i, "__ERROR__"]);
    }

    const numberOfDimensions = (
      temporaryRepresentation.get("dimensions") as Collection
    )?.items?.length;

    // if no dimensions, this block is skipped.
    for (let i = 0; i < numberOfDimensions; i++) {
      const dimension = temporaryRepresentation.getIn([
        "dimensions",
        i,
      ]) as YAMLMap;

      if (dimension.has("__ERROR__"))
        temporaryRepresentation.deleteIn(["dimensions", i, "__ERROR__"]);
    }

    this.internalYAML = temporaryRepresentation.toString({
      collectionStyle: "block",
    });

    // Update svelte store
    this.updateStore(this);

    if (shouldUpdateRuntime) {
      // Update Runtime
      this.updateRuntime(this.internalYAML);
    }
  }

  getMetricKey<K extends keyof MetricsConfig>(key: K): MetricsConfig[K] {
    return this.internalRepresentation[key];
  }

  updateMetricsParams(params: Partial<MetricsParams>) {
    for (const param in params) {
      this.internalRepresentationDocument.set(param, params[param]);
    }
    this.regenerateInternalYAML();
  }

  updateErrors(errors: Array<V1ReconcileError>) {
    // set errors for measures and dimensions
    for (const error of errors) {
      const index = Number(error.propertyPath[1]);
      switch (error.propertyPath[0]) {
        case "Measures":
          this.internalRepresentationDocument.setIn(
            ["measures", index, "__ERROR__"],
            error.message
          );
          break;
        case "Dimensions":
          this.internalRepresentationDocument.setIn(
            ["dimensions", index, "__ERROR__"],
            error.message
          );
          break;
      }
    }

    this.regenerateInternalYAML(false);
  }

  // MEASURE METHODS
  getMeasures() {
    return this.internalRepresentation.measures;
  }

  addNewMeasure() {
    const measureNode = this.internalRepresentationDocument.createNode({
      label: "",
      expression: "",
      description: "",
      format_preset: "humanize",
      __GUID__: guidGenerator(),
    });

    this.internalRepresentationDocument.addIn(["measures"], measureNode);
    this.regenerateInternalYAML();
  }

  deleteMeasure(index: number) {
    this.internalRepresentationDocument.deleteIn(["measures", index]);
    this.regenerateInternalYAML();
  }

  updateMeasure(index: number, key: string, change) {
    this.internalRepresentationDocument.setIn(["measures", index, key], change);
    this.regenerateInternalYAML();
  }

  // DIMENSIONS METHODS
  getDimensions() {
    return this.internalRepresentation.dimensions;
  }

  addNewDimension() {
    const dimensionNode = this.internalRepresentationDocument.createNode({
      label: "",
      property: "",
      description: "",
    });

    this.internalRepresentationDocument.addIn(["dimensions"], dimensionNode);
    this.regenerateInternalYAML();
  }

  updateDimension(index: number, key: string, change) {
    this.internalRepresentationDocument.setIn(
      ["dimensions", index, key],
      change
    );
    this.regenerateInternalYAML();
  }

  deleteDimension(index: number) {
    this.internalRepresentationDocument.deleteIn(["dimensions", index]);
    this.regenerateInternalYAML();
  }
}

export function createInternalRepresentation(yamlString, updateRuntime) {
  const metricRep = new MetricsInternalRepresentation(
    yamlString,
    updateRuntime
  );

  return readable(metricRep, (set) => {
    metricRep.bindStore((instance) => {
      set(instance);
    });
  });
}

const capitalize = (s) => s && s[0].toUpperCase() + s.slice(1);

export function initBlankDashboardYAML(dashboardName: string) {
  const metricsTemplate = `
# Visit https://docs.rilldata.com/references/project-files to learn more about Rill project files.

display_name: ""
model: ""
default_time_range: ""
default_time_grain: ""
time_grains: []
timeseries: ""
measures: []
dimensions: []
`;
  const template = parseDocument(metricsTemplate);
  template.set("display_name", dashboardName);
  return template.toString();
}

export function addQuickMetricsToDashboardYAML(yaml: string, model: V1Model) {
  const doc = parseDocument(yaml);
  doc.set("model", model.name);

  const timestampColumns = selectTimestampColumnFromSchema(model?.schema);
  if (timestampColumns?.length) {
    doc.set("timeseries", timestampColumns[0]);
  } else {
    doc.set("timeseries", "");
  }

  const measureNode = doc.createNode({
    label: "Total records",
    expression: "count(*)",
    description: "Total number of records present",
    format_preset: "humanize",
  });
  doc.set("measures", [measureNode]);

  const fields = model.schema.fields;
  const diemensionSeq = fields
    .filter((field) => {
      return CATEGORICALS.has(field.type.code);
    })
    .map((field) => {
      return {
        label: capitalize(field.name),
        property: field.name,
        description: "",
      };
    });

  const dimensionNode = doc.createNode(diemensionSeq);
  doc.set("dimensions", dimensionNode);

  return doc.toString({ collectionStyle: "block" });
}
