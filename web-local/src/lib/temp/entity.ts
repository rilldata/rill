export enum EntityType {
  Table = "Table",
  Model = "Model",
  Application = "Application",
  MetricsDefinition = "MetricsDefinition",
  MeasureDefinition = "MeasureDefinition",
  DimensionDefinition = "DimensionDefinition",
  MetricsExplorer = "MetricsExplorer",
}

export enum StateType {
  Persistent = "Persistent",
  Derived = "Derived",
}

export interface EntityRecord {
  id: string;
  type: EntityType;
  lastUpdated: number;
}

export enum EntityStatus {
  Idle,
  Running,
  Error,

  Importing,
  Validating,
  Profiling,
  Exporting,
}
