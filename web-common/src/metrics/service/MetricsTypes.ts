import { EntityType } from "@rilldata/web-common/features/entity-management/types";

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
  Modal = "modal",
}

export enum MetricsEventScreenName {
  Source = "source",
  Model = "model",
  Dashboard = "dashboard",
  MetricsDefinition = "metrics-definition",
  CLI = "cli",
  Splash = "splash",
  Home = "home",
  Organization = "organization",
  Project = "project",
  Report = "report",
  ReportExport = "report-export",
  Alert = "alert",
  Unknown = "unknown",
}

export const ScreenToEntityMap = {
  [MetricsEventScreenName.Source]: EntityType.Table,
  [MetricsEventScreenName.Model]: EntityType.Model,
  [MetricsEventScreenName.Dashboard]: EntityType.MetricsDefinition,
  [MetricsEventScreenName.MetricsDefinition]: EntityType.MetricsDefinition,
  [MetricsEventScreenName.Home]: EntityType.Application,
  [MetricsEventScreenName.Splash]: EntityType.Application,
};

export interface ActiveEvent extends MetricsEvent {
  event_type: "active";
  duration_sec: number;
  total_in_focus: number;
}
