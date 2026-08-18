import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
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
  // Base fields required by the telemetry service. For details, see rill/runtime/pkg/activity/README.md.
  event_id: string;
  event_time: string;
  event_type: string;
  event_name: string;
  // The URL at the moment the event was fired.
  // This is what attributes an event to the dashboard or canvas the user was on, since the resource
  // name only exists in the URL.
  // Note that the query string is a snapshot, not a history: it reflects the dashboard state at the
  // instant of this event, so what it tells you depends on when the event fires. Page views are
  // deduped per path and view mode, so their filters and time range are whatever the page loaded with.
  page_url: string;
  // Legacy:
  event_datetime: number;
}

export enum MetricsEventSpace {
  RightPanel = "right-panel",
  Workspace = "workspace",
  LeftPanel = "left-panel",
  Modal = "modal",
}

export enum MetricsEventScreenName {
  Table = "table",
  Source = "source",
  Model = "model",
  Dashboard = "dashboard",
  MetricsDefinition = "metrics-definition",
  Chart = "chart",
  Canvas = "canvas",
  Connector = "connector",
  CLI = "cli",
  Splash = "splash",
  Home = "home",
  Organization = "organization",
  Project = "project",
  Report = "report",
  ReportExport = "report-export",
  Alert = "alert",
  Unknown = "unknown",
  Explore = "explore",
  Pivot = "pivot",
}

export const ScreenToEntityMap = {
  [MetricsEventScreenName.Source]: EntityType.Table,
  [MetricsEventScreenName.Model]: EntityType.Model,
  [MetricsEventScreenName.Dashboard]: EntityType.MetricsDefinition,
  [MetricsEventScreenName.MetricsDefinition]: EntityType.MetricsDefinition,
  [MetricsEventScreenName.Home]: EntityType.Application,
  [MetricsEventScreenName.Splash]: EntityType.Application,
};

export const ResourceKindToScreenMap: Partial<
  Record<ResourceKind, MetricsEventScreenName>
> = {
  [ResourceKind.Source]: MetricsEventScreenName.Source,
  [ResourceKind.Model]: MetricsEventScreenName.Model,
  [ResourceKind.MetricsView]: MetricsEventScreenName.Dashboard,
  [ResourceKind.Component]: MetricsEventScreenName.Chart,
  [ResourceKind.Canvas]: MetricsEventScreenName.Canvas,
  [ResourceKind.Connector]: MetricsEventScreenName.Connector,
};

export interface ActiveEvent extends MetricsEvent {
  event_type: "active";
  duration_sec: number;
  total_in_focus: number;
}
