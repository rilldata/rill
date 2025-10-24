import ApiIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
import TableIcon from "@rilldata/web-common/components/icons/TableIcon.svelte";
import ThemeIcon from "@rilldata/web-common/components/icons/ThemeIcon.svelte";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { Code2Icon } from "lucide-svelte";
import ConnectorIcon from "../../components/icons/ConnectorIcon.svelte";
import MetricsViewIcon from "../../components/icons/MetricsViewIcon.svelte";

export const resourceIconMapping = {
  [ResourceKind.Source]: TableIcon,
  [ResourceKind.Connector]: ConnectorIcon,
  [ResourceKind.Model]: Code2Icon,
  [ResourceKind.MetricsView]: MetricsViewIcon,
  [ResourceKind.Explore]: ExploreIcon,
  [ResourceKind.API]: ApiIcon,
  [ResourceKind.Component]: Chart,
  [ResourceKind.Canvas]: CanvasIcon,
  [ResourceKind.Theme]: ThemeIcon,
  [ResourceKind.Report]: ReportIcon,
  [ResourceKind.Alert]: AlertIcon,
};

export const resourceColorMapping = {
  [ResourceKind.Source]: "#059669",
  [ResourceKind.Connector]: "#6B7280",
  [ResourceKind.Model]: "#0891B2",
  [ResourceKind.MetricsView]: "#7C3AED",
  [ResourceKind.Explore]: "#4736F5",
  [ResourceKind.API]: "#EA580C",
  [ResourceKind.Component]: "#65A30D",
  [ResourceKind.Canvas]: "#008FD4",
  [ResourceKind.Theme]: "#DB2777",
  [ResourceKind.Report]: "#0284C7",
  [ResourceKind.Alert]: "#F59E0B",
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
