export enum AddDataStep {
  SelectConnector,
  CreateConnector,
  CreateModel,
  ExploreConnector,
  Import,
  Done,
}

export type AddDataConfig = {
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
  | SelectConnectorStep
  | CreateConnectorStep
  | CreateModelStep
  | ExploreConnectorStep
  | ImportAddDataStep
  | DoneAddDataStep;

/**
 * Individual steps for strong typing
 */

type SelectConnectorStep = {
  step: AddDataStep.SelectConnector;
};

type CreateConnectorStep = {
  step: AddDataStep.CreateConnector;
  schema: string;
};

type CreateModelStep = {
  step: AddDataStep.CreateModel;
  schema: string;
  connector: string;
};

type ExploreConnectorStep = {
  step: AddDataStep.ExploreConnector;
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
