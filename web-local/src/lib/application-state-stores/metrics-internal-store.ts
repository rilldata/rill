import type { V1Model } from "@rilldata/web-common/runtime-client";
import { guidGenerator } from "@rilldata/web-local/lib/util/guid";
import { readable, Subscriber } from "svelte/store";
import { Document, ParsedNode, parseDocument, YAMLMap } from "yaml";
import type { Collection } from "yaml/dist/nodes/Collection";
import { CATEGORICALS, TIMESTAMPS } from "../duckdb-data-types";

export const metricsTemplate = `
display_name: "Sample Dashboard"
description: "a description that appears in the UI"

# model
#optional to declare this, otherwise it is the model.sql file in the same directory
from: ""

# populate with the first datetime type in the OBT
timeseries: ""

# default to opionated option around estimated timegrain,
# first in order is default time grain
timegrains:
  - "DAY"
# the timegrain that users will see when they first visit the dashboard.
default_timegrain:
  - "DAY"

# measures
# measures are presented in the order that they are written in this file.
measures: []

# dimensions
# dimensions are presented in the order that they are written in this file.
dimensions: []
`;

export interface MetricsConfig {
  display_name: string;
  description: string;
  timeseries: string;
  timegrains?: Array<string>;
  default_timegrain?: Array<string>;
  from: string;
  measures: MeasureEntity[];
  dimensions: DimensionEntity[];
}
export interface MeasureEntity {
  label?: string;
  expression?: string;
  description?: string;
  format_preset?: string;
  visible?: boolean;
  __GUID__?: string;
}
export interface DimensionEntity {
  label?: string;
  property?: string;
  description?: string;
  visible?: boolean;
  expression?: string;
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

    this.updateRuntime = updateRuntime;
  }

  bindStore(updateStore: Subscriber<MetricsInternalRepresentation>) {
    this.updateStore = updateStore;
  }

  decorateInternalRepresentation(yamlString: string) {
    const internalRepresentationDoc = parseDocument(yamlString);
    const numberOfMeasures = (
      internalRepresentationDoc.get("measures") as Collection
    ).items.length;

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

  regenerateInternalYAML() {
    const temporaryRepresentation = this.internalRepresentationDocument.clone();
    const numberOfMeasures = (
      temporaryRepresentation.get("measures") as Collection
    ).items.length;

    Array(numberOfMeasures)
      .fill(0)
      .map((_, i) => {
        const measure = temporaryRepresentation.getIn([
          "measures",
          i,
        ]) as YAMLMap;

        if (measure.has("__GUID__"))
          temporaryRepresentation.deleteIn(["measures", i, "__GUID__"]);

        if (measure.has("__ERROR__"))
          temporaryRepresentation.deleteIn(["measures", i, "__ERROR__"]);
      });

    this.internalYAML = temporaryRepresentation.toString({
      collectionStyle: "block",
    });
    this.internalRepresentation = this.internalRepresentationDocument.toJSON();

    // Update svelte store
    this.updateStore(this);

    // Update Runtime
    this.updateRuntime(this.internalYAML);
  }

  getMetricKey(key: keyof MetricsConfig) {
    return this.internalRepresentation[key];
  }

  updateMetricKey(key: keyof MetricsConfig, value) {
    this.internalRepresentationDocument.set(key, value);
    this.regenerateInternalYAML();
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
      visible: true,
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
      expression: "",
      visible: true,
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

export function generateMeasuresAndDimension(
  model: V1Model,
  timeseries?: string
) {
  const fields = model.schema.fields;

  const template = parseDocument(metricsTemplate);
  template.set("from", model.name);

  if (timeseries) {
    template.set("timeseries", timeseries);
  } else {
    const timestampColumns = model.schema.fields
      .filter((column) => TIMESTAMPS.has(column.type.code as string))
      .map((column) => column.name);

    template.set("timeseries", timestampColumns[0]);
  }
  const measureNode = template.createNode({
    label: "Total records",
    expression: "count(*)",
    description: "Total number of records present",
    format_preset: "humanize",
    visible: true,
  });
  template.addIn(["measures"], measureNode);

  const diemensionSeq = fields
    .filter((field) => {
      return CATEGORICALS.has(field.type.code);
    })
    .map((field) => {
      return {
        label: "",
        property: field.name,
        description: "",
        expression: "",
        visible: true,
      };
    });

  const dimensionNode = template.createNode(diemensionSeq);
  template.set("dimensions", dimensionNode);

  return template.toString({ collectionStyle: "block" });
}
