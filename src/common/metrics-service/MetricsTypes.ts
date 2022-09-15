import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface CommonFields {
  app_name: string;
  install_id: string;
  build_id: string;
  version: string;
  is_dev: boolean;
  project_id: string;
}

export interface CommonUserFields {
  locale: string;
  browser: string;
  os: string;
  device_model: string;
}

export interface MetricsEvent extends CommonFields, CommonUserFields {
  event_datetime: number;
  event_type: string;
}

export enum MetricsEventSpace {
  RightPanel = "right-panel",
  Workspace = "workspace",
  LeftPanel = "left-panel",
}

export enum MetricsEventScreenName {
  Source = "source",
  Model = "model",
  Dashboard = "dashboard",
  MetricsDefinition = "metrics-definition",
}

export const EntityTypeToScreenMap = {
  [EntityType.Table]: MetricsEventScreenName.Source,
  [EntityType.Model]: MetricsEventScreenName.Model,
  [EntityType.Application]: MetricsEventScreenName.Source,
  [EntityType.MetricsDefinition]: MetricsEventScreenName.MetricsDefinition,
  [EntityType.MeasureDefinition]: MetricsEventScreenName.MetricsDefinition,
  [EntityType.MetricsExplorer]: MetricsEventScreenName.Dashboard,
  [EntityType.DimensionDefinition]: MetricsEventScreenName.Dashboard,
};

export interface ActiveEvent extends MetricsEvent {
  event_type: "active";
  duration_sec: number;
  total_in_focus: number;
}
