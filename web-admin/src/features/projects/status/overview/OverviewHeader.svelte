<script lang="ts">
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useProjectDeployment, useRuntimeVersion } from "../selectors";
  import { formatEnvironmentName, getStatusDotClass } from "../display-utils";
  import ProjectClone from "../project-information/ProjectClone.svelte";

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  $: projectDeployment = useProjectDeployment(organization, project);
  $: deployment = $projectDeployment.data;
  $: deploymentStatus =
    deployment?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;
  $: deploymentEnvironment = formatEnvironmentName(deployment?.environment);

  $: proj = createAdminServiceGetProject(organization, project);
  $: projectData = $proj.data?.project;
  $: gitRemote = projectData?.gitRemote;
  $: managedGitId = projectData?.managedGitId;
  $: primaryBranch = projectData?.primaryBranch;
  $: isGithubConnected = !!gitRemote && !managedGitId;

  $: githubLastSynced = useGithubLastSynced(instanceId);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    instanceId,
    organization,
    project,
  );
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;

  $: runtimeVersionQuery = useRuntimeVersion();
  $: version = $runtimeVersionQuery.data?.version ?? "";
</script>

<div class="flex flex-col gap-2">
  <div class="flex items-center justify-between gap-4 flex-wrap">
    <div class="flex items-center gap-2">
      <h2 class="text-lg font-semibold text-fg-primary">Overview</h2>
      <span class="status-dot {getStatusDotClass(deploymentStatus)}"></span>
    </div>
    <div class="flex items-center gap-3">
      {#if version}
        <span class="text-xs font-mono text-fg-secondary">{version}</span>
      {/if}
      <ProjectClone {organization} {project} />
    </div>
  </div>
  <div class="flex items-center gap-2 text-sm text-fg-secondary">
    <span>{deploymentEnvironment}</span>
    {#if isGithubConnected}
      <span class="font-mono font-semibold">{primaryBranch}</span>
    {/if}
    {#if lastUpdated}
      <span class="text-gray-300">|</span>
      <span>
        Last synced {lastUpdated.toLocaleString(undefined, {
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
        })}
      </span>
    {/if}
  </div>
</div>

<style lang="postcss">
  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }
</style>
