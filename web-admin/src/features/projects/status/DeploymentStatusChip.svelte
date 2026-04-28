<script lang="ts">
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { Chip } from "@rilldata/web-common/components/chip";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/Cancel.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircleNew.svelte";
  import { CircleDashedIcon } from "lucide-svelte";
  import { getStatusLabel } from "@rilldata/web-admin/features/projects/status/display-utils.ts";

  let {
    organization,
    project,
  }: {
    organization: string;
    project: string;
  } = $props();

  let projectQuery = $derived(
    createAdminServiceGetProject(organization, project),
  );
  let deployment = $derived($projectQuery.data?.deployment);
  let label = $derived(
    deployment?.status ? getStatusLabel(deployment.status) : "Unpublished",
  );

  const DeploymentStatus: Record<
    V1DeploymentStatus,
    {
      type: "dimension" | "special" | "amber" | "measure" | "time";
      icon: any;
    }
  > = {
    [V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED]: {
      type: "dimension",
      icon: CircleDashedIcon,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING]: {
      type: "time",
      icon: LoadingSpinner,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING]: {
      type: "time",
      icon: LoadingSpinner,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING]: {
      type: "dimension",
      icon: CheckCircle,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED]: {
      type: "amber",
      icon: CancelCircle,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING]: {
      type: "amber",
      icon: LoadingSpinner,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED]: {
      type: "amber",
      icon: CancelCircle,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING]: {
      type: "amber",
      icon: LoadingSpinner,
    },
    [V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED]: {
      type: "amber",
      icon: CancelCircle,
    },
  };
  let status = $derived(
    DeploymentStatus[
      deployment?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
    ],
  );
  const IconComponent = $derived(status.icon);
</script>

<Chip gray={!deployment} type={status.type} compact readOnly>
  <div slot="body" class="flex flex-row items-center gap-1">
    <IconComponent this={status.icon} size="12px" />
    <span>{label}</span>
  </div>
</Chip>
