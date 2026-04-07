import {
  type MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes.ts";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";

export enum AddDataStep {
  // Used purely to transition from Init to one of the other steps.
  // This is used to start the import flow from the middle by selecting schema/connector directly.
  Init = "init",
  SelectConnector = "select-connector",
  CreateConnector = "create-connector",
  CreateModel = "create-model",
  ExploreConnector = "explore-connector",
  Import = "import",
  Done = "done",
}

export type AddDataConfig = {
  welcomeScreen?: boolean;
  importOnly?: boolean;

  // Telemetry related config
  medium: BehaviourEventMedium;
  space: MetricsEventSpace;
  screen: MetricsEventScreenName;
};

export enum ImportDataStep {
  Init = "init",
  CreateModel = "create-model",
  CreateMetricsView = "create-metrics-view",
  CreateDashboard = "create-dashboard",
  Done = "done",
}
export const ImportDataStepsOrder: Record<ImportDataStep, number> = {
  [ImportDataStep.Init]: 0,
  [ImportDataStep.CreateModel]: 1,
  [ImportDataStep.CreateMetricsView]: 2,
  [ImportDataStep.CreateDashboard]: 3,
  [ImportDataStep.Done]: 4,
};

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
  // Used only for firing telemetry events
  schema: string;
  importStep: ImportDataStep;
  currentFilePath?: string;
  config: ImportStepConfig;
};
