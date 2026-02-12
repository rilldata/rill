<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceListDeployments,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    createRuntimeServiceGetInstance,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    useProjectDeployment,
    useRuntimeVersion,
    useResources,
  } from "../selectors";
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
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

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

  // Deployment counts
  $: allDeployments = createAdminServiceListDeployments(
    organization,
    project,
    {},
  );
  // Will Use these when Rill Cloud Editing is supported. This will allow users to see how many deployments they have and in which environments. See L188
  // $: deploymentCount = $allDeployments.data?.deployments?.length ?? 0;
  // $: prodCount =
  //   $allDeployments.data?.deployments?.filter((d) => d.environment === "prod")
  //     .length ?? 0;
  // $: devCount =
  //   $allDeployments.data?.deployments?.filter((d) => d.environment === "dev")
  //     .length ?? 0;

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

  // Resources
  $: resources = useResources(instanceId);
  $: allResources = $resources.data?.resources ?? [];

  const displayKinds = [
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
    ResourceKind.Alert,
    ResourceKind.Report,
    ResourceKind.API,
    ResourceKind.Connector,
  ];

  $: resourceCounts = countByKind(allResources);

  function countByKind(
    res: V1Resource[],
  ): { kind: string; label: string; count: number }[] {
    const counts = new Map<string, number>();
    for (const r of res) {
      const kind = r.meta?.name?.kind;
      if (kind) counts.set(kind, (counts.get(kind) ?? 0) + 1);
    }
    return displayKinds
      .filter((kind) => (counts.get(kind) ?? 0) > 0)
      .map((kind) => ({
        kind,
        label: prettyResourceKind(kind),
        count: counts.get(kind) ?? 0,
      }));
  }
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
        <!-- Hide counts for now since we only show primary deployment, which is usually prod. Can add back if we show multiple deployments in the future. -->
        <!-- <span class="text-fg-tertiary text-xs ml-1">
          ({deploymentCount} total · {prodCount} prod · {devCount} dev)
        </span> -->
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

  {#if resourceCounts.length > 0}
    <div class="resource-chips">
      {#each resourceCounts as { kind, label, count }}
        <a href="{basePage}/resources?kind={kind}" class="resource-chip">
          {#if resourceIconMapping[kind]}
            <svelte:component this={resourceIconMapping[kind]} size="12px" />
          {/if}
          <span class="font-medium">{count}</span>
          <span class="text-fg-secondary">{label}{count !== 1 ? "s" : ""}</span>
        </a>
      {/each}
    </div>
  {/if}
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
    @apply flex items-center py-2 ;
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
  .resource-chips {
    @apply flex flex-wrap gap-2 mt-4 pt-4 border-t border-border;
  }
  .resource-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle;
  }
  .resource-chip:hover {
    @apply border-primary-500 text-primary-600;
  }
</style>
