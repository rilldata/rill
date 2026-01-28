import ApiIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
import TableIcon from "@rilldata/web-common/components/icons/TableIcon.svelte";
import ThemeIcon from "@rilldata/web-common/components/icons/ThemeIcon.svelte";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import ConnectorIcon from "../../components/icons/ConnectorIcon.svelte";
import MetricsViewIcon from "../../components/icons/MetricsViewIcon.svelte";
import ModelIcon from "@rilldata/web-common/components/icons/ModelIcon.svelte";
import File from "@rilldata/web-common/components/icons/File.svelte";
import SettingsIcon from "@rilldata/web-common/components/icons/SettingsIcon.svelte";

export const resourceIconMapping = {
  // Source is deprecated and merged with Model - use Model icon
  [ResourceKind.Source]: ModelIcon,
  [ResourceKind.Connector]: ConnectorIcon,
  [ResourceKind.Model]: ModelIcon,
  [ResourceKind.MetricsView]: MetricsViewIcon,
  [ResourceKind.Explore]: ExploreIcon,
  [ResourceKind.API]: ApiIcon,
  [ResourceKind.Component]: Chart,
  [ResourceKind.Canvas]: CanvasIcon,
  [ResourceKind.Theme]: ThemeIcon,
  [ResourceKind.Report]: ReportIcon,
  [ResourceKind.Alert]: AlertIcon,
};

export const resourceLabelMapping = {
  [ResourceKind.Source]: "Source",
  [ResourceKind.Connector]: "Connector",
  [ResourceKind.Model]: "Model",
  [ResourceKind.MetricsView]: "Metrics View",
  [ResourceKind.Explore]: "Explore",
  [ResourceKind.API]: "API",
  [ResourceKind.Component]: "Component",
  [ResourceKind.Canvas]: "Canvas",
  [ResourceKind.Theme]: "Theme",
  [ResourceKind.Report]: "Report",
  [ResourceKind.Alert]: "Alert",
};

export const resourceShorthandMapping = {
  [ResourceKind.Source]: "source",
  [ResourceKind.Connector]: "connector",
  [ResourceKind.Model]: "model",
  [ResourceKind.MetricsView]: "metrics",
  [ResourceKind.Explore]: "explore",
  [ResourceKind.API]: "api",
  [ResourceKind.Component]: "component",
  [ResourceKind.Canvas]: "canvas",
  [ResourceKind.Theme]: "theme",
  [ResourceKind.Report]: "report",
  [ResourceKind.Alert]: "alert",
};

export function getIconComponent(
  kind: ResourceKind | undefined,
  filePath: string,
) {
  return kind
    ? resourceIconMapping[kind]
    : filePath === "/.env" || filePath === "/rill.yaml"
      ? SettingsIcon
      : File;
}
