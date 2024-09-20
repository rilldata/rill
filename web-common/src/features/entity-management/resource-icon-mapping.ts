import ApiIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
import MetricsExplorerIcon from "@rilldata/web-common/components/icons/MetricsExplorerIcon.svelte";
import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
import ThemeIcon from "@rilldata/web-common/components/icons/ThemeIcon.svelte";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { Code2Icon, Database } from "lucide-svelte";

export const resourceIconMapping = {
  [ResourceKind.Source]: Database,
  [ResourceKind.Connector]: Database,
  [ResourceKind.Model]: Code2Icon,
  [ResourceKind.MetricsView]: MetricsExplorerIcon,
  [ResourceKind.API]: ApiIcon,
  [ResourceKind.Component]: Chart,
  [ResourceKind.Canvas]: CanvasIcon,
  [ResourceKind.Theme]: ThemeIcon,
  [ResourceKind.Report]: ReportIcon,
  [ResourceKind.Alert]: AlertCircleOutline,
};
