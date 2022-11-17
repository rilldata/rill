import { guidGenerator } from "@rilldata/web-local/lib/util/guid";
import { Document, ParsedNode, parseDocument } from "yaml";

export interface MetricsConfig {
  display_name: string;
  description: string;
  timeseries: string;
  timegrains?: Array<string>;
  default_timegrain?: Array<string>;
  model_path: string;
  measures: MeasureEntity[][];
  dimensions: DimensionEntity[][];
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
  yamlString: string;
  internalYAML: string;
  internalRepresentation: MetricsConfig;
  internalRepresentationDocument: Document.Parsed<ParsedNode>;
  measures: MeasureEntity[] = [];
  dimensions: DimensionEntity[];

  constructor(yamlString: string) {
    this.yamlString = yamlString;
    this.internalRepresentation = this.decorateInternalRepresentation(
      yamlString
    ) as MetricsConfig;
  }

  decorateInternalRepresentation(yamlString: string) {
    const internalRepresentation = parseDocument(yamlString);
    const measures = internalRepresentation.get(
      "measures"
    ) as MeasureEntity[][];

    measures.forEach((measure) => {
      const reducedMeasures: MeasureEntity = {};

      for (let i = 0; i < measure.length; i++) {
        Object.assign(reducedMeasures, measure[i]);
      }
      reducedMeasures.__GUID__ = guidGenerator();

      this.measures.push(reducedMeasures);
    });

    this.internalRepresentationDocument = internalRepresentation;

    return internalRepresentation.toJSON();
  }

  regenerateInternalYAML() {
    this.internalYAML = this.internalRepresentationDocument.toString();
  }
}
