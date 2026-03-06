export enum AddDataStep {
  Select,
  Olap,
  Connector,
  Source,
  Explorer,
  Import,
  Done,
}

export type AddDataConfig = {
  instanceId: string;
  importOnly?: boolean;
};

export type AddDataTransitionArgs = {
  schema?: string;
  connector?: string;
  importConfig?: ImportAddDataStepConfig;
};

export enum ImportDataStep {
  Init,
  CreateModel,
  CreateMetricsView,
  CreateExplore,
  Done,
}

export type AddDataState =
  | SelectAddDataStep
  | OlapAddDataStep
  | ConnectorAddDataStep
  | SourceAddDataStep
  | ExplorerAddDataStep
  | ImportAddDataStep
  | DoneAddDataStep;

/**
 * Individual steps for strong typing
 */

type SelectAddDataStep = {
  step: AddDataStep.Select;
};

type OlapAddDataStep = {
  step: AddDataStep.Olap;
  schema: string; // Forward to the connector/source step
};

type ConnectorAddDataStep = {
  step: AddDataStep.Connector;
  schema: string;
};

type SourceAddDataStep = {
  step: AddDataStep.Source;
  schema: string;
  connector: string;
};

type ExplorerAddDataStep = {
  step: AddDataStep.Explorer;
  schema: string;
  connector: string;
};

type DoneAddDataStep = {
  step: AddDataStep.Done;
  finalPath: string;
};

// Import data step and types

export type ImportAddDataStepConfig = {
  importSteps: ImportDataStep[];
  source: string;
  sourceSchema: string;
  sourceDatabase: string;
  connector: string;
  yaml: string;
  envBlob: string | null;
};

export type ImportAddDataStep = {
  step: AddDataStep.Import;
  importStep: ImportStep;
  currentFilePath: string;
  config: ImportAddDataStepConfig;
};

type ImportStep =
  | InitImportStep
  | CreateModelImportStep
  | CreateMetricsViewImportStep
  | CreateExploreImportStep
  | DoneImportStep;

type InitImportStep = {
  step: ImportDataStep.Init;
};

type CreateModelImportStep = {
  step: ImportDataStep.CreateModel;
  source: string;
  connector: string;
  yaml: string;
  envBlob: string | null;
};

type CreateMetricsViewImportStep = {
  step: ImportDataStep.CreateMetricsView;
  source: string;
  sourceSchema: string;
  sourceDatabase: string;
  connector: string;
};

type CreateExploreImportStep = {
  step: ImportDataStep.CreateExplore;
  metricsViewFilePath: string;
};

type DoneImportStep = {
  step: ImportDataStep.Done;
};
