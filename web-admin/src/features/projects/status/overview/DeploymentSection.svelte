<script lang="ts">
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useProjectDeployment, useRuntimeVersion } from "../selectors";
  import {
    formatEnvironmentName,
    formatConnectorName,
    getStatusDotClass,
    getStatusLabel,
  } from "../display-utils";
  import ProjectClone from "./ProjectClone.svelte";

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  // Deployment
  $: projectDeployment = useProjectDeployment(organization, project);
  $: deployment = $projectDeployment.data;
  $: deploymentStatus =
    deployment?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;

  // Project
  $: proj = createAdminServiceGetProject(organization, project);
  $: projectData = $proj.data?.project;
  $: primaryBranch = projectData?.primaryBranch;
  // Last synced
  $: githubLastSynced = useGithubLastSynced(instanceId);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    instanceId,
    organization,
    project,
  );
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;

  // Runtime
  $: runtimeVersionQuery = useRuntimeVersion();
  $: version = $runtimeVersionQuery.data?.version ?? "";

  // Connectors
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.olapConnector,
  );
  $: aiConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.aiConnector,
  );
</script>

<section class="section">
  <div class="section-header">
    <h3 class="section-title">Deployment</h3>
    <ProjectClone {organization} {project} />
  </div>

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

    {#if primaryBranch}
      <div class="info-row">
        <span class="info-label">Branch</span>
        <span class="info-value font-mono text-xs">{primaryBranch}</span>
      </div>
    {/if}

    {#if lastUpdated}
      <div class="info-row">
        <span class="info-label">Last synced</span>
        <span class="info-value">
          {lastUpdated.toLocaleString(undefined, {
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "numeric",
          })}
        </span>
      </div>
    {/if}

    {#if version}
      <div class="info-row">
        <span class="info-label">Runtime</span>
        <span class="info-value font-mono text-xs">{version}</span>
      </div>
    {/if}

    <div class="info-row">
      <span class="info-label">OLAP Engine</span>
      <span class="info-value">
        {olapConnector ? formatConnectorName(olapConnector.type) : "—"}
        {#if olapConnector && (olapConnector.provision || olapConnector.type !== "duckdb")}
          <span class="text-fg-tertiary text-xs ml-1">
            ({olapConnector.provision ? "Rill-managed" : "Self-managed"})
          </span>
        {/if}
      </span>
    </div>

    <div class="info-row">
      <span class="info-label">AI</span>
      <span class="info-value">
        {#if aiConnector}
          {formatConnectorName(aiConnector.name)}
        {:else}
          Rill Managed
        {/if}
      </span>
    </div>
  </div>
</section>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
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
</style>
