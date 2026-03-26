<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceTriggerRedeploy,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useProjectDeployment, useRuntimeVersion } from "../selectors";
  import {
    formatEnvironmentName,
    formatConnectorName,
    getOlapEngineLabel,
    getStatusDotClass,
    getStatusLabel,
  } from "../display-utils";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import ProjectClone from "./ProjectClone.svelte";
  import OverviewCard from "@rilldata/web-common/features/projects/status/overview/OverviewCard.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;
  export let project: string;
  export let branch: string | undefined = undefined;

  const runtimeClient = useRuntimeClient();
  const redeployMutation = createAdminServiceTriggerRedeploy();

  // Deployment
  $: projectDeployment = useProjectDeployment(organization, project, branch);
  $: deployment = $projectDeployment.data;
  $: deploymentStatus =
    deployment?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;

  // Project
  $: proj = createAdminServiceGetProject(
    organization,
    project,
    branch ? { branch } : undefined,
  );
  $: projectData = $proj.data?.project;
  $: primaryBranch = projectData?.primaryBranch;
  // Last synced
  $: githubLastSynced = useGithubLastSynced(runtimeClient);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    runtimeClient,
    organization,
    project,
  );
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;

  // Runtime
  $: runtimeVersionQuery = useRuntimeVersion(runtimeClient);
  $: version = $runtimeVersionQuery.data?.version?.match(/v[\d.]+/)?.[0] ?? "";

  // Connectors — sensitive: true is needed to read projectConnectors (OLAP/AI connector types)
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  // Repo — only shown when the user connected their own GitHub
  $: githubUrl = projectData?.gitRemote
    ? getGitUrlFromRemote(projectData.gitRemote)
    : "";
  $: isGithubConnected =
    !!projectData?.gitRemote && !projectData?.managedGitId && !!githubUrl;

  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.olapConnector,
  );
  $: olapEngineLabel = getOlapEngineLabel(olapConnector);
  $: aiConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.aiConnector,
  );

  async function handleSync() {
    try {
      await $redeployMutation.mutateAsync({
        data: { org: organization, project, deploymentId: deployment?.id },
      });
      eventBus.emit("notification", {
        type: "success",
        message: "Sync triggered",
      });
    } catch {
      eventBus.emit("notification", {
        type: "error",
        message: "Failed to trigger sync",
      });
    }
  }
</script>

<OverviewCard title="Deployment">
  <ProjectClone
    slot="header-right"
    {organization}
    {project}
    gitRemote={projectData?.gitRemote}
    managedGitId={projectData?.managedGitId}
  />

  <div class="info-grid">
    <div class="info-row">
      <span class="info-label">Status</span>
      <span class="info-value flex items-center gap-2">
        <span class="status-dot {getStatusDotClass(deploymentStatus)}"></span>
        {getStatusLabel(deploymentStatus)}
      </span>
    </div>

    <div class="info-row">
      <span class="info-label">Environment</span>
      <span class="info-value">
        {formatEnvironmentName(deployment?.environment)}
      </span>
    </div>

    {#if isGithubConnected}
      <div class="info-row">
        <span class="info-label">Repo</span>
        <span class="info-value">
          <a
            href={githubUrl}
            target="_blank"
            rel="noopener noreferrer"
            class="repo-link"
          >
            {githubUrl.replace("https://github.com/", "")}
          </a>
        </span>
      </div>
    {/if}

    {#if isGithubConnected && (deployment?.branch || primaryBranch)}
      <div class="info-row">
        <span class="info-label">Branch</span>
        <span class="info-value">{deployment?.branch || primaryBranch}</span>
      </div>
    {/if}

    {#if lastUpdated}
      <div class="info-row">
        <span class="info-label">Last synced</span>
        <span class="info-value flex items-center gap-2">
          {lastUpdated.toLocaleString(undefined, {
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "numeric",
          })}
          <button
            class="sync-button"
            onclick={handleSync}
            disabled={$redeployMutation.isPending}
          >
            {$redeployMutation.isPending ? "Syncing..." : "Sync now"}
          </button>
        </span>
      </div>
    {/if}

    {#if version}
      <div class="info-row">
        <span class="info-label">Runtime</span>
        <span class="info-value">{version}</span>
      </div>
    {/if}

    <div class="info-row">
      <span class="info-label">OLAP Engine</span>
      <span class="info-value">{olapEngineLabel}</span>
    </div>

    <div class="info-row">
      <span class="info-label">AI Connector</span>
      <span class="info-value">
        {#if aiConnector && aiConnector.name !== "admin"}
          {formatConnectorName(aiConnector.type)}
          <span class="text-fg-tertiary text-xs ml-1">({aiConnector.name})</span
          >
        {:else}
          Rill Managed
        {/if}
      </span>
    </div>
  </div>
</OverviewCard>

<style lang="postcss">
  .info-grid {
    @apply flex flex-col;
  }
  .info-row {
    @apply flex items-center py-2;
  }
  .info-row:last-child {
    @apply border-b-0;
  }
  .info-label {
    @apply text-sm text-fg-secondary w-32 shrink-0;
  }
  .info-value {
    @apply text-sm text-fg-primary;
  }
  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }
  .repo-link {
    @apply text-primary-500 text-sm;
  }
  .repo-link:hover {
    @apply underline;
  }
  .sync-button {
    @apply text-xs text-primary-500 hover:text-primary-700 disabled:text-fg-muted cursor-pointer disabled:cursor-default;
  }
</style>
