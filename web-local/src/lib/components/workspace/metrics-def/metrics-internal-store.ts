import { guidGenerator } from "@rilldata/web-local/lib/util/guid";
import { get, readable, Subscriber, writable } from "svelte/store";
import { Document, ParsedNode, parseDocument, YAMLMap } from "yaml";
import type { Collection } from "yaml/dist/nodes/Collection";

export interface MetricsConfig {
  display_name: string;
  description: string;
  timeseries: string;
  timegrains?: Array<string>;
  default_timegrain?: Array<string>;
  model_path: string;
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

  // Svelte method to set store value
  set?: Subscriber<MetricsInternalRepresentation>;

  updateStore: (instance: MetricsInternalRepresentation) => void;

  constructor(yamlString: string) {
    this.internalRepresentation = this.decorateInternalRepresentation(
      yamlString
    ) as MetricsConfig;
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
  }

  getMetricKey(key) {
    return this.internalRepresentation[key];
  }

  updateMetricKey(key, value) {
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

    this.updateStore(this);
  }

  deleteMeasure(index: number) {
    this.internalRepresentationDocument.deleteIn(["measures", index]);
    this.regenerateInternalYAML();
    this.updateStore(this);
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
    this.updateStore(this);
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
    this.updateStore(this);
  }
}

export function createInternalRepresentation(yamlString) {
  const metricRep = new MetricsInternalRepresentation(yamlString);

  const store = writable(metricRep);
  metricRep.bindStore((instance) => {
    store.update((_) => instance);

    console.log(
      "measures in store",
      get(store).internalRepresentation.measures
    );
  });

  return store;

  // return readable(metricRep, (set) => {
  //   metricRep.bindStore((instance) => {
  //     console.log("Instance", instance);
  //     set(instance);
  //   });
  // });
}
