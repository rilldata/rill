<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { isTeamPlan } from "@rilldata/web-admin/features/billing/plans/utils";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
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

  export let organization: string;
  export let project: string;

  const runtimeClient = useRuntimeClient();

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
  $: dataSizeBytes = $instanceQuery.data?.dataSizeBytes;

  // Plan-based storage cap (from root layout data)
  const TEAM_STORAGE_CAP = 10 * 1024 * 1024 * 1024; // 10GB
  $: planName = $page.data?.organization?.billingPlanName ?? "";
  $: storageCap = isTeamPlan(planName) ? TEAM_STORAGE_CAP : 0;

  // Fill percentage for the usage pill (0–100)
  $: usagePercent = (() => {
    const bytes = Number(dataSizeBytes ?? 0);
    if (!storageCap) return bytes > 0 ? 100 : 0;
    return Math.min(Math.round((bytes / storageCap) * 100), 100);
  })();
  $: isOverCap = storageCap > 0 && Number(dataSizeBytes ?? 0) >= storageCap;

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

    {#if isGithubConnected && primaryBranch}
      <div class="info-row">
        <span class="info-label">Branch</span>
        <span class="info-value">{primaryBranch}</span>
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

    {#if !olapConnector || olapConnector.provision}
      <div class="info-row">
        <span class="info-label">Data usage</span>
        <span class="info-value flex items-center gap-2">
          {#if dataSizeBytes}
            <a
              href="/{organization}/{project}/-/status/tables"
              class="usage-pill-link"
              aria-label="Data usage"
            >
              <span class="usage-pill">
                <span
                  class="usage-pill-fill"
                  class:over-cap={isOverCap}
                  style:width="{usagePercent}%"
                ></span>
              </span>
            </a>
            <span class="text-xs text-fg-secondary whitespace-nowrap">
              {formatMemorySize(Number(dataSizeBytes))}{#if storageCap}
                / {formatMemorySize(storageCap)}{/if}
            </span>
          {:else}
            —
          {/if}
        </span>
      </div>
    {/if}
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
  .usage-pill-link {
    @apply no-underline;
  }
  .usage-pill {
    @apply w-24 h-2.5 rounded-full bg-surface-subtle overflow-hidden inline-block;
  }
  .usage-pill-fill {
    @apply h-full rounded-full bg-primary-500 block transition-all;
  }
  .repo-link {
    @apply text-primary-500 text-sm;
  }
  .repo-link:hover {
    @apply underline;
  }
</style>
