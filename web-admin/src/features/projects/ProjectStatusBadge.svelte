<script lang="ts">
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import {
    CheckCircle2,
    CircleDashed,
    XCircle,
    type Icon,
  } from "lucide-svelte";
  import type { ComponentType } from "svelte";

  export let deploymentStatus: V1DeploymentStatus | undefined;
  export let isPublic: boolean;
  export let hasDeployment: boolean;

  type Variant = "ready" | "live" | "error" | "unpublished" | "pending";

  $: variant = getVariant(deploymentStatus, isPublic, hasDeployment);

  function getVariant(
    status: V1DeploymentStatus | undefined,
    pub: boolean,
    deployed: boolean,
  ): Variant {
    if (status === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED) return "error";
    if (
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING ||
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING
    )
      return "pending";
    if (status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING)
      return pub ? "live" : "ready";
    if (!deployed) return "unpublished";
    return "unpublished";
  }

  const config: Record<
    Variant,
    { label: string; icon: ComponentType<Icon>; classes: string }
  > = {
    ready: {
      label: "Ready",
      icon: CheckCircle2,
      classes: "bg-blue-50 border-blue-500 text-blue-500",
    },
    live: {
      label: "Live",
      icon: CheckCircle2,
      classes: "bg-green-50 border-green-500 text-green-500",
    },
    error: {
      label: "Error",
      icon: XCircle,
      classes: "bg-red-50 border-red-600 text-red-600",
    },
    unpublished: {
      label: "Unpublished",
      icon: CircleDashed,
      classes: "bg-surface-muted border-border text-fg-tertiary",
    },
    pending: {
      label: "Pending",
      icon: CircleDashed,
      classes: "bg-surface-muted border-border text-fg-tertiary",
    },
  };

  $: ({ label, icon, classes } = config[variant]);
</script>

<span
  class="inline-flex items-center gap-x-1 h-7 px-2.5 rounded-2xl border text-sm font-medium leading-5 shadow-xs {classes}"
>
  <svelte:component this={icon} size="12" strokeWidth={2.5} />
  {label}
</span>
