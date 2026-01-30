<script lang="ts">
  import {
    V1DeploymentStatus,
    createAdminServiceGetProject,
  } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import GithubConnectionDialog from "@rilldata/web-admin/features/projects/github/GithubConnectionDialog.svelte";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Brain from "@rilldata/web-common/components/icons/Brain.svelte";
  import TableIcon from "@rilldata/web-common/components/icons/TableIcon.svelte";
  import Code from "@rilldata/web-common/components/icons/Code.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    getRepoNameFromGitRemote,
    getGitUrlFromRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useProjectDeployment, useRuntimeVersion } from "./selectors";
  import ProjectClone from "./ProjectClone.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  // Project data
  $: proj = createAdminServiceGetProject(organization, project);
  $: projectData = $proj.data?.project;
  $: gitRemote = projectData?.gitRemote;
  $: managedGitId = projectData?.managedGitId;
  $: primaryBranch = projectData?.primaryBranch;
  $: subpath = projectData?.subpath;
  $: isGithubConnected = !!gitRemote;
  $: isManagedGit = !!managedGitId;
  $: repoName = gitRemote ? getRepoNameFromGitRemote(gitRemote) : "";

  // Deployment data
  $: projectDeployment = useProjectDeployment(organization, project);
  $: ({ data: deployment, isLoading: deploymentLoading } = $projectDeployment);
  $: deploymentStatus =
    deployment?.status || V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;
  $: deploymentEnvironment = formatEnvironmentName(deployment?.environment);

  function formatEnvironmentName(env: string | undefined): string {
    if (!env) return "Production";
    const lower = env.toLowerCase();
    if (lower === "prod" || lower === "production") return "Production";
    if (lower === "dev" || lower === "development") return "Development";
    if (lower === "stage" || lower === "staging") return "Staging";
    // Capitalize first letter for other environments
    return env.charAt(0).toUpperCase() + env.slice(1);
  }

  // Simple status indicator (green/yellow/red/gray)
  function getStatusDotClass(status: V1DeploymentStatus): string {
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
        return "bg-green-500"; // Green - Ready
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING:
        return "bg-yellow-500"; // Yellow - In progress
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
        return "bg-red-500"; // Red - Error
      default:
        return "bg-gray-400"; // Gray - Not deployed
    }
  }

  function getStatusLabel(status: V1DeploymentStatus): string {
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
        return "Ready";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
        return "Pending";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
        return "Updating";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING:
        return "Stopping";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING:
        return "Deleting";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
        return "Error";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED:
        return "Stopped";
      case V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED:
        return "Deleted";
      default:
        return "Not deployed";
    }
  }

  // Version
  $: runtimeVersionQuery = useRuntimeVersion();
  $: version = $runtimeVersionQuery.data?.version || "—";

  // Instance data for OLAP and AI connectors
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: false,
  });
  $: olapConnector = $instanceQuery.data?.instance?.olapConnector;
  $: aiConnector = $instanceQuery.data?.instance?.aiConnector;

  // Last synced
  $: githubLastSynced = useGithubLastSynced(instanceId);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    instanceId,
    organization,
    project,
  );
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;

  // Format connector name for display
  function formatConnectorName(connector: string | undefined): string {
    if (!connector) return "—";
    // Capitalize first letter and clean up common names
    if (connector === "duckdb") return "DuckDB";
    if (connector === "clickhouse") return "ClickHouse";
    if (connector === "druid") return "Druid";
    if (connector === "pinot") return "Pinot";
    if (connector === "openai") return "OpenAI";
    if (connector === "anthropic") return "Anthropic";
    return connector.charAt(0).toUpperCase() + connector.slice(1);
  }

  function handleEditInWeb() {
    // TODO: Navigate to edit route
    console.log("Edit in web");
  }
</script>

<!-- Header row with status and version (outside the box) -->
<div class="header">
  <div class="header-left">
    <h2 class="title">Project Status</h2>
    {#if deploymentLoading}
      <Spinner status={EntityStatus.Running} size="16px" />
    {:else}
      <div class="deployment-info">
        <span class="deployment-env">{deploymentEnvironment}:</span>
        <Tooltip distance={8}>
          <span class="status-dot {getStatusDotClass(deploymentStatus)}"></span>
          <TooltipContent slot="tooltip-content">
            <p class="tooltip-text">{getStatusLabel(deploymentStatus)}</p>
          </TooltipContent>
        </Tooltip>
        {#if deployment?.statusMessage}
          <span class="status-message">— {deployment.statusMessage}</span>
        {/if}
      </div>
    {/if}
  </div>
  <div class="version">
    {version}
  </div>
</div>

<!-- Info grid (inside the box) -->
<div class="info-box">
  <div class="info-grid">
    <!-- GitHub -->
    <div class="info-cell github-cell">
      <div class="cell-header">
        <Github size="16px" color="#6b7280" />
        <span class="cell-label">GitHub</span>
        <Tooltip distance={8}>
          <a
            href="https://docs.rilldata.com/developers/deploy/deploy-dashboard/github-101"
            target="_blank"
            rel="noreferrer noopener"
            class="info-link"
          >
            <InfoCircle size="14px" color="#9ca3af" />
          </a>
          <TooltipContent slot="tooltip-content">
            <p class="tooltip-text">
              Unlock BI-as-code with GitHub-backed collaboration, version
              control, and approval workflows.
            </p>
          </TooltipContent>
        </Tooltip>
      </div>
      <div class="cell-content">
        {#if isGithubConnected && !isManagedGit}
          <div class="flex items-center gap-1">
            <Github size="16px" />
            <a
              href={getGitUrlFromRemote(gitRemote)}
              class="repo-link"
              target="_blank"
              rel="noreferrer noopener"
            >
              {repoName}
            </a>
          </div>
          {#if subpath}
            <div class="github-detail">
              <span class="detail-label">subpath</span>: /{subpath}
            </div>
          {/if}
          <div class="github-detail">
            <span class="detail-label">branch</span>: {primaryBranch}
          </div>
          {#if lastUpdated}
            <span class="synced-text">
              Synced {lastUpdated.toLocaleString(undefined, {
                month: "short",
                day: "numeric",
                hour: "numeric",
                minute: "numeric",
              })}
            </span>
          {/if}
        {:else}
          <GithubConnectionDialog {organization} {project} />
        {/if}
      </div>
    </div>

    <!-- OLAP Engine -->
    <div class="info-cell">
      <div class="cell-header">
        <span class="icon-muted"><TableIcon size="16px" /></span>
        <span class="cell-label">OLAP Engine</span>
        <Tooltip distance={8}>
          <a
            href="https://docs.rilldata.com/developers/build/connectors/olap"
            target="_blank"
            rel="noreferrer noopener"
            class="info-link"
          >
            <InfoCircle size="14px" color="#9ca3af" />
          </a>
          <TooltipContent slot="tooltip-content">
            <p class="tooltip-text">Learn about supported OLAP engines.</p>
          </TooltipContent>
        </Tooltip>
      </div>
      <div class="cell-content">
        {#if olapConnector && olapConnector !== "duckdb"}
          <span class="connector-name"
            >{formatConnectorName(olapConnector)}</span
          >
        {:else}
          <span class="connector-name">DuckDB</span>
          <span class="cell-meta">Rill Managed</span>
        {/if}
      </div>
    </div>

    <!-- AI Connector -->
    <div class="info-cell">
      <div class="cell-header">
        <Brain size="16" color="#6b7280" />
        <span class="cell-label">AI</span>
        <Tooltip distance={8}>
          <a
            href="https://docs.rilldata.com/developers/build/connectors/data-source/openai"
            target="_blank"
            rel="noreferrer noopener"
            class="info-link"
          >
            <InfoCircle size="14px" color="#9ca3af" />
          </a>
          <TooltipContent slot="tooltip-content">
            <p class="tooltip-text">
              Configure AI connectors for your project.
            </p>
          </TooltipContent>
        </Tooltip>
      </div>
      <div class="cell-content">
        {#if aiConnector}
          <span class="connector-name">{formatConnectorName(aiConnector)}</span>
        {:else}
          <span class="connector-name">OpenAI</span>
          <span class="cell-meta">Rill Managed</span>
        {/if}
      </div>
    </div>

    <!-- Local Development -->
    <div class="info-cell">
      <div class="cell-header">
        <span class="icon-muted"><Code size="16px" /></span>
        <span class="cell-label">Local Development</span>
        <Tooltip distance={8}>
          <a
            href="https://docs.rilldata.com/developers/guides/clone-a-project"
            target="_blank"
            rel="noreferrer noopener"
            class="info-link"
          >
            <InfoCircle size="14px" color="#9ca3af" />
          </a>
          <TooltipContent slot="tooltip-content">
            <p class="tooltip-text">Clone this project to develop locally.</p>
          </TooltipContent>
        </Tooltip>
      </div>
      <div class="cell-content">
        <div class="button-group">
          <ProjectClone {organization} {project} />
          <!-- <Button type="secondary" onClick={handleEditInWeb}>
            Edit in Web
          </Button> -->
        </div>
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .header {
    @apply flex items-center justify-between mb-3;
  }

  .header-left {
    @apply flex items-center gap-3;
  }

  .title {
    @apply text-lg font-semibold text-gray-900;
  }

  .status-dot {
    @apply w-2 h-2 rounded-full;
  }

  .status-text {
    @apply text-sm font-medium text-gray-700;
  }

  .version {
    @apply text-sm font-mono text-gray-500;
  }

  .deployment-info {
    @apply flex items-center gap-2;
  }

  .deployment-env {
    @apply text-sm font-medium text-gray-700;
  }

  .status-message {
    @apply text-sm text-gray-600;
    @apply max-w-md truncate;
  }

  .info-box {
    @apply p-4 bg-white border border-gray-200 rounded-lg;
  }

  .info-grid {
    @apply grid grid-cols-4 gap-6;
  }

  .info-cell {
    @apply flex flex-col gap-1.5;
  }

  .github-cell {
    @apply gap-2;
  }

  .cell-header {
    @apply flex items-center gap-1.5;
  }

  .cell-label {
    @apply text-xs font-medium text-gray-500 uppercase tracking-wide;
  }

  .cell-content {
    @apply flex flex-col gap-1;
  }

  .repo-link {
    @apply text-sm font-semibold text-gray-900;
    @apply font-mono truncate;
  }

  .repo-link:hover {
    @apply text-primary-600;
  }

  .github-detail {
    @apply text-sm text-gray-700;
  }

  .detail-label {
    @apply font-mono text-gray-500;
  }

  .synced-text {
    @apply text-xs text-gray-500;
  }

  .github-promo {
    @apply text-sm text-gray-500 my-1;
  }

  .connector-name {
    @apply text-sm font-medium text-gray-900;
  }

  .cell-meta {
    @apply text-xs text-gray-500;
    @apply flex items-center gap-1;
  }

  .button-group {
    @apply flex gap-2;
  }

  .icon-muted {
    @apply text-gray-500;
  }

  .icon-muted :global(svg path) {
    fill: currentColor;
  }

  .info-link {
    @apply flex items-center;
  }

  .info-link:hover :global(svg) {
    color: #6b7280;
  }

  .tooltip-text {
    @apply text-sm max-w-[200px];
  }
</style>
