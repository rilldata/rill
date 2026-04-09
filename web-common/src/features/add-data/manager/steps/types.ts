export enum AddDataStep {
  // Used purely to transition from Init to one of the other steps.
  // This is used to start the import flow from the middle by selecting schema/connector directly.
  Init,
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
  skipNavigation?: boolean;
};

export enum ImportDataStep {
  Init,
  CreateModel,
  CreateMetricsView,
  CreateDashboard,
  Done,
}

export type AddDataState =
  | InitConnectorStep
  | SelectConnectorStep
  | CreateConnectorStep
  | CreateModelStep
  | ExploreConnectorStep
  | ImportAddDataStep
  | DoneAddDataStep;
export type AddDataStepWithSchema =
  | CreateConnectorStep
  | CreateModelStep
  | ExploreConnectorStep;
export type AddDataStepWithConnector = CreateModelStep | ExploreConnectorStep;

/**
 * Individual steps for strong typing
 */

type InitConnectorStep = {
  step: AddDataStep.Init;
};

type SelectConnectorStep = {
  step: AddDataStep.SelectConnector;
};

export type CreateConnectorStep = {
  step: AddDataStep.CreateConnector;
  schema: string;
  // Generated ID used to fetch cached info in the connector form.
  // Used to reuse form state when user comes back to this step.
  connectorId: string;
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
