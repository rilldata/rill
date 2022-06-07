// FIXME what are the correct root types below?
// these types at the top are placeholders to begin thinking about the state shape

// is the EntityRecordId the correct id to use to lookup a Model and get info about it columns etc?
type SourceModelEntityId = number | string;

// are model columns stored by id of any kind, or only name?
type ModelColumnIdOrName = number | string;
// or whatever we're usng for ids
type UUID = string;

enum ValidationState {
  OK = "OK",
  WARNING = "WARNING",
  ERROR = "ERROR",
}

export type MeasureDefinition = {
  // mandatory user defined metadata
  expression: string;
  // optional user defined metadata
  sqlName?: string;
  label?: string;
  description?: string;
  // internal state for rendering etc
  id: UUID;
  expressionIsValid: ValidationState;
  sqlNameIsValid: ValidationState;
  sparkLineId: UUID;
};

export type DimensionDefinition = {
  // mandatory user defined data
  dimensionColumn: string;
  // optional user defined metadata
  sqlName?: string;
  nameSingle?: string;
  namePlural?: string;
  description?: string;
  // internal state for rendering etc
  id: UUID;
  dimensionIsValid: ValidationState;
  sqlNameIsValid: ValidationState;
  summaryPlotId: UUID; // Want to reuse summary cardnality plots used elsewhere. How are those stored?
};

export type MetricsDefinition = {
  metricDefinitionId: UUID;
  metricDefLabel: string;
  sourceModelId: SourceModelEntityId | undefined;
  timeDimension: ModelColumnIdOrName | undefined;
  measures: MeasureDefinition[];
  dimensions: DimensionDefinition[];
};

export type MetricsDefinitionsSlice = {
  defs: { [id: UUID]: MetricsDefinition };
  defsCounter: number;
  selectedDefId?: UUID;
};
