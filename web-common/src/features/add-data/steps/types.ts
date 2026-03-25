export enum AddDataStep {
  SelectConnector,
  CreateConnector,
  CreateModel,
  ExploreConnector,
  Import,
  Done,
}

export type AddDataConfig = {
  welcomeScreen?: boolean;
  importOnly?: boolean;
};

export type AddDataTransitionArgs = {
  schema?: string;
  connector?: string;
  importConfig?: ImportStepConfig;
  connectorFormValues?: Record<string, unknown>;
};

export enum ImportDataStep {
  Init,
  CreateModel,
  CreateMetricsView,
  CreateExplore,
  CreateCanvas,
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

export type CreateConnectorStep = {
  step: AddDataStep.CreateConnector;
  schema: string;
};

export type CreateModelStep = {
  step: AddDataStep.CreateModel;
  schema: string;
  connector: string;
  connectorFormValues: Record<string, unknown>;
};

export type ExploreConnectorStep = {
  step: AddDataStep.ExploreConnector;
  schema: string;
  connector: string;
};

type DoneAddDataStep = {
  step: AddDataStep.Done;
  finalPath: string;
};

// Import data step and types

export type ImportStepConfig = {
  importSteps: ImportDataStep[];
  connector: string;
  importFrom: ImportFromConfig;
  importTo: ImportToConfig;
  envBlob: string | null;
};

export type ImportFromConfig =
  | {
      from: "table";
      table: string;
      schema: string;
      database: string;
    }
  | {
      from: "sql";
      sql: string;
    }
  | {
      from: "yaml";
      yaml: string;
    };

// Generated names for consistency across retries
export type ImportToConfig = {
  modelName?: string;
  modelPath?: string;
  metricsViewName?: string;
  metricsViewPath?: string;
  exploreName?: string;
  explorePath?: string;
  canvasName?: string;
  canvasPath?: string;
};

export type ImportAddDataStep = {
  step: AddDataStep.Import;
  importStep: ImportDataStep;
  currentFilePath?: string;
  config: ImportStepConfig;
};
