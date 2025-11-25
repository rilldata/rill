import { V1DeploymentStatus } from "@rilldata/web-admin/client";
import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
import InfoCircleFilled from "@rilldata/web-common/components/icons/InfoCircleFilled.svelte";
import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

export type StatusDisplay = {
  icon: any; // SvelteComponent
  iconProps?: {
    [key: string]: unknown;
  };
  text?: string;
  textClass?: string;
  wrapperClass?: string;
};

export const deploymentChipDisplays: Record<V1DeploymentStatus, StatusDisplay> =
  {
    [V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED]: {
      icon: InfoCircleFilled,
      iconProps: { className: "text-indigo-600 hover:text-indigo-500" },
      text: "Not deployed",
      textClass: "text-indigo-600",
      wrapperClass: "bg-indigo-50 border-indigo-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED]: {
      icon: InfoCircleFilled,
      iconProps: { className: "text-indigo-600 hover:text-indigo-500" },
      text: "Not deployed",
      textClass: "text-indigo-600",
      wrapperClass: "bg-indigo-50 border-indigo-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING]: {
      icon: Spinner,
      iconProps: {
        bg: "linear-gradient(90deg, #22D3EE -0.5%, #6366F1 98.5%)",
        className: "text-purple-600 hover:text-purple-500",
        status: EntityStatus.Running,
      },
      text: "Pending",
      textClass: "text-purple-600",
      wrapperClass: "bg-purple-50 border-purple-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING]: {
      icon: Spinner,
      iconProps: {
        bg: "linear-gradient(90deg, #22D3EE -0.5%, #6366F1 98.5%)",
        className: "text-purple-600 hover:text-purple-500",
        status: EntityStatus.Running,
      },
      text: "Updating",
      textClass: "text-purple-600",
      wrapperClass: "bg-purple-50 border-purple-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING]: {
      icon: Spinner,
      iconProps: {
        bg: "linear-gradient(90deg, #22D3EE -0.5%, #6366F1 98.5%)",
        className: "text-purple-600 hover:text-purple-500",
        status: EntityStatus.Running,
      },
      text: "Stopping",
      textClass: "text-purple-600",
      wrapperClass: "bg-purple-50 border-purple-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING]: {
      icon: Spinner,
      iconProps: {
        bg: "linear-gradient(90deg, #22D3EE -0.5%, #6366F1 98.5%)",
        className: "text-purple-600 hover:text-purple-500",
        status: EntityStatus.Running,
      },
      text: "Deleting",
      textClass: "text-purple-600",
      wrapperClass: "bg-purple-50 border-purple-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED]: {
      icon: InfoCircleFilled,
      iconProps: { className: "text-indigo-600 hover:text-indigo-500" },
      text: "Deleted",
      textClass: "text-indigo-600",
      wrapperClass: "bg-indigo-50 border-indigo-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED]: {
      icon: CancelCircle,
      iconProps: { className: "text-red-600 hover:text-red-500" },
      text: "Error",
      textClass: "text-red-600",
      wrapperClass: "bg-red-50 border-red-300",
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING]: {
      icon: CheckCircle,
      iconProps: { className: "text-primary-600 hover:text-primary-500" },
      text: "Ready",
      textClass: "text-primary-600",
      wrapperClass: "bg-primary-50 border-primary-300",
    },
  };

export function getResourceKindTagColor(kind: string) {
  switch (kind) {
    case "rill.runtime.v1.MetricsView":
      return "blue";
    case "rill.runtime.v1.Model":
      return "green";
    case "rill.runtime.v1.Report":
      return "orange";
    case "rill.runtime.v1.Source":
      return "purple";
    case "rill.runtime.v1.Theme":
      return "magenta";
    default:
      return "gray";
  }
}
