import { guidGenerator } from "@rilldata/web-local/lib/util/guid";
import { Document, ParsedNode, parseDocument } from "yaml";
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

  constructor(yamlString: string) {
    this.internalRepresentation = this.decorateInternalRepresentation(
      yamlString
    ) as MetricsConfig;
  }

  decorateInternalRepresentation(yamlString: string) {
    const internalRepresentation = parseDocument(yamlString);
    const measures = (internalRepresentation.get("measures") as Collection)
      .items;

    measures.forEach((measure: MeasureEntity) => {
      measure.__GUID__ = guidGenerator();
    });
    this.internalRepresentationDocument = internalRepresentation;

    return internalRepresentation.toJSON();
  }

  regenerateInternalYAML() {
    const temporaryRepresentation = this.internalRepresentationDocument.clone();
    const measures = (temporaryRepresentation.get("measures") as Collection)
      .items;

    measures.forEach((measure: MeasureEntity) => {
      delete measure.__GUID__;
    });

    this.internalYAML = temporaryRepresentation.toString();
    this.internalRepresentation = this.internalRepresentationDocument.toJSON();
  }

  getModel() {
    return this.internalRepresentation.model_path;
  }

  updateModel(model_path) {
    this.internalRepresentationDocument.set("model_path", model_path);
    this.regenerateInternalYAML();
  }

  getTitle() {
    return this.internalRepresentation.display_name;
  }

  updateTitle(title) {
    this.internalRepresentationDocument.set("display_name", title);
    this.regenerateInternalYAML();
  }

  // MEASURE METHODS
  getMeasures() {
    return this.internalRepresentation.measures;
  }

  addNewMeasure() {
    this.internalRepresentationDocument.addIn(["measures"], {
      label: "",
      expression: "",
      description: "",
      format_preset: "",
      visible: true,
      __GUID__: guidGenerator(),
    });

    this.regenerateInternalYAML();
  }

  deleteMeasure(index: number) {
    this.internalRepresentationDocument.deleteIn(["measures", index]);
    this.regenerateInternalYAML();
  }

  updateMeasure(index: number, key: string, change) {
    const prevMeasure = this.internalRepresentationDocument.getIn([
      "measures",
      index,
    ]);
    prevMeasure[key] = change;
    this.internalRepresentationDocument.setIn(["measures", index], prevMeasure);

    this.regenerateInternalYAML();
  }

  // DIMENSIONS METHODS
  getDimensions() {
    return this.internalRepresentation.dimensions;
  }

  addNewDimension() {
    this.internalRepresentationDocument.addIn(["dimensions"], {
      label: "",
      property: "",
      description: "",
      expression: "",
      visible: true,
    });

    this.regenerateInternalYAML();
  }

  updateDimension(index: number, key: string, change) {
    const prevDimension = this.internalRepresentationDocument.getIn([
      "dimensions",
      index,
    ]);
    prevDimension[key] = change;
    this.internalRepresentationDocument.setIn(
      ["dimensions", index],
      prevDimension
    );

    this.regenerateInternalYAML();
  }

  deleteDimension(index: number) {
    this.internalRepresentationDocument.deleteIn(["dimensions", index]);
    this.regenerateInternalYAML();
  }
}
