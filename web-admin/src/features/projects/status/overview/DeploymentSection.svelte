<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetProjectVariables,
    createAdminServiceGetBillingSubscription,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    isTrialPlan,
    isEnterprisePlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useProjectDeployment, useRuntimeVersion } from "../selectors";
  import {
    formatEnvironmentName,
    formatConnectorName,
    getStatusDotClass,
    getStatusLabel,
  } from "../display-utils";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ProjectClone from "./ProjectClone.svelte";
  import ManageSlotsModal from "./ManageSlotsModal.svelte";
  import ClickHouseCloudKeyModal from "./ClickHouseCloudKeyModal.svelte";
  import ClickHouseCloudDetailsModal from "./ClickHouseCloudDetailsModal.svelte";
  import OverviewCard from "./OverviewCard.svelte";
  import { onMount } from "svelte";

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
  // Repo — only shown when the user connected their own GitHub
  $: githubUrl = projectData?.gitRemote
    ? getGitUrlFromRemote(projectData.gitRemote)
    : "";
  $: isGithubConnected =
    !!projectData?.gitRemote && !projectData?.managedGitId && !!githubUrl;

  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.olapConnector,
  );
  $: aiConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.aiConnector,
  );

  // Slots
  $: currentSlots = Number(projectData?.prodSlots) || 0;
  // Live Connect only when a non-DuckDB connector explicitly has provision=false.
  // DuckDB is always Rill-managed (including local dev where provision=false).
  $: isRillManaged =
    !olapConnector ||
    olapConnector.type === "duckdb" ||
    olapConnector.provision === true;
  $: canManage = $proj.data?.projectPermissions?.manageProject ?? false;
  let slotsModalOpen = false;

  // Billing plan detection
  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: planName = $subscriptionQuery?.data?.subscription?.plan?.name ?? "";
  $: isTrial = isTrialPlan(planName);
  $: isEnterprise = planName !== "" && isEnterprisePlan(planName);

  // Slot usage breakdown (dev edit modes coming soon; each consumes 1 slot)
  $: prodSlots = currentSlots; // today all slots go to prod
  $: devSlots = 0; // will increase when dev edit modes are active
  $: usedSlots = prodSlots + devSlots;

  // ClickHouse Cloud detection: prefer backend flag, fall back to host/DSN check
  $: isClickHouseCloud =
    olapConnector?.type === "clickhouse" &&
    (olapConnector?.config?.is_clickhouse_cloud === true ||
      (olapConnector?.config?.host as string)
        ?.toLowerCase()
        .endsWith(".clickhouse.cloud") ||
      (olapConnector?.config?.dsn as string)
        ?.toLowerCase()
        .includes(".clickhouse.cloud"));

  // CHC service details from the runtime (populated after API key is saved and polling runs)
  $: cloudServiceName = olapConnector?.config?.cloud_service_name as
    | string
    | undefined;
  $: cloudStatus = olapConnector?.config?.cloud_status as string | undefined;
  $: cloudProvider = olapConnector?.config?.cloud_provider as
    | string
    | undefined;
  $: cloudRegion = olapConnector?.config?.cloud_region as string | undefined;
  $: cloudMinMemory = olapConnector?.config?.cloud_min_memory_gb as
    | number
    | undefined;
  $: cloudMaxMemory = olapConnector?.config?.cloud_max_memory_gb as
    | number
    | undefined;
  $: cloudReplicas = olapConnector?.config?.cloud_num_replicas as
    | number
    | undefined;
  $: chcAutoScaleAnnotation =
    projectData?.annotations?.["rill.dev/chc-auto-scaled-slots"] === "true";
  $: isChcHibernated =
    cloudStatus === "idle" ||
    cloudStatus === "stopped" ||
    cloudStatus === "stopping" ||
    // If annotation is set and status is unknown, treat as hibernated
    (chcAutoScaleAnnotation && !cloudStatus);
  // CHC is running again but slots haven't been restored yet
  $: isChcRestoring =
    !isChcHibernated &&
    cloudStatus === "running" &&
    chcAutoScaleAnnotation;


  let chcDetailsModalOpen = false;

  $: projectVariablesQuery = isClickHouseCloud
    ? createAdminServiceGetProjectVariables(organization, project, {
        environment: "prod",
      })
    : undefined;
  $: hasCloudApiKey = $projectVariablesQuery?.data?.variables?.some(
    (v) => v.name === "CLICKHOUSE_CLOUD_API_KEY_ID",
  );

  // "Remind me later" dismisses for this browser session; re-opens on new visits
  $: dismissKey = `chc-key-dismissed:${organization}/${project}`;
  let chcDismissedThisSession = false;
  onMount(() => {
    chcDismissedThisSession =
      sessionStorage.getItem(`chc-key-dismissed:${organization}/${project}`) ===
      "true";
  });

  function handleChcDismiss() {
    sessionStorage.setItem(dismissKey, "true");
    chcDismissedThisSession = true;
  }

  $: shouldPromptChcKey =
    isClickHouseCloud &&
    hasCloudApiKey === false &&
    !chcDismissedThisSession &&
    canManage;
  let chcKeyModalOpen = false;
  $: if (shouldPromptChcKey) {
    chcKeyModalOpen = true;
  }

  // If CHC key exists but no slots are configured, open the required slots modal once on page load
  let slotsPromptChecked = false;
  $: if (
    !slotsPromptChecked &&
    isClickHouseCloud &&
    hasCloudApiKey === true &&
    currentSlots === 0 &&
    canManage
  ) {
    slotsPromptChecked = true;
    slotsRequiredMode = true;
    slotsModalOpen = true;
  }

  // Use the resolved host from the runtime (handles DSN parsing server-side)
  $: connectorHost = olapConnector?.config?.resolved_host as string | undefined;

  $: console.log(
    "[CHC] connectorHost:",
    connectorHost,
    "olapConnector config:",
    olapConnector?.config,
  );

  // After CHC key is saved, open the slots modal in required mode
  let slotsRequiredMode = false;

  // Persist detected memory in sessionStorage so it survives modal reopens and navigations
  const chcMemoryKey = `chc-memory:${organization}/${project}`;
  let chcDetectedMemoryGb: number | undefined = (() => {
    const stored = sessionStorage.getItem(chcMemoryKey);
    return stored ? parseFloat(stored) : undefined;
  })();

  function handleChcKeySaved(memoryGb?: number) {
    console.log(
      "[CHC] Key saved, detectedMemoryGb:",
      memoryGb,
      "canManage:",
      canManage,
    );
    if (memoryGb && memoryGb > 0) {
      chcDetectedMemoryGb = memoryGb;
      sessionStorage.setItem(chcMemoryKey, String(memoryGb));
    }
    if (canManage) {
      slotsRequiredMode = true;
      slotsModalOpen = true;
    }
  }
  // Reset required mode when slots modal closes
  $: if (!slotsModalOpen) slotsRequiredMode = false;

  // CHC auto-scaling: detect when slots were auto-reduced due to CHC hibernation
  $: isChcAutoScaled =
    projectData?.annotations?.["rill.dev/chc-auto-scaled-slots"] === "true" &&
    currentSlots === 1;


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
        {#if isChcHibernated}
          <Tooltip distance={8} location="top">
            <span class="flex items-center gap-2">
              <span class="status-dot bg-yellow-500"></span>
              Unhealthy
            </span>
            <TooltipContent slot="tooltip-content">
              ClickHouse Cloud is {cloudStatus === "idle"
                ? "waking up"
                : cloudStatus === "stopping"
                  ? "stopping"
                  : "hibernated"}
            </TooltipContent>
          </Tooltip>
        {:else if isChcRestoring}
          <Tooltip distance={8} location="top">
            <span class="flex items-center gap-2">
              <span class="status-dot bg-yellow-500"></span>
              Preparing project
            </span>
            <TooltipContent slot="tooltip-content">
              ClickHouse Cloud is back online. Restoring slots.
            </TooltipContent>
          </Tooltip>
        {:else}
          <span class="status-dot {getStatusDotClass(deploymentStatus)}"></span>
          {getStatusLabel(deploymentStatus)}
        {/if}
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
      <span class="info-value flex items-center gap-2">
        {olapConnector ? formatConnectorName(olapConnector.type) : "DuckDB"}
        {#if olapConnector && (olapConnector.provision || olapConnector.type !== "duckdb")}
          <span class="text-fg-tertiary text-xs">
            ({olapConnector.provision
              ? "Rill-managed"
              : isClickHouseCloud
                ? "ClickHouse Cloud"
                : "Self-managed"})
          </span>
        {/if}
        {#if isClickHouseCloud && hasCloudApiKey}
          <button
            class="manage-slots-btn"
            on:click={() => (chcDetailsModalOpen = true)}
          >
            View Details
          </button>
        {:else if isClickHouseCloud && !hasCloudApiKey && canManage}
          <button
            class="manage-slots-btn"
            on:click={() => (chcKeyModalOpen = true)}
          >
            Connect to ClickHouse Cloud
          </button>
        {/if}
      </span>
    </div>

    <div class="info-row">
      <span class="info-label">AI</span>
      <span class="info-value">
        {#if aiConnector}
          {formatConnectorName(aiConnector.type)}
          <span class="text-fg-tertiary text-xs ml-1">({aiConnector.name})</span
          >
        {:else}
          Rill Managed
        {/if}
      </span>
    </div>

    {#if !isEnterprise}
      <div class="info-row">
        <span class="info-label">Slots</span>
        <span class="info-value flex items-center gap-3">
          {#if currentSlots > 0}
            <div class="slots-pill">
              <div
                class="slots-fill-prod"
                style="width: {(prodSlots / currentSlots) * 100}%"
              ></div>
              <div
                class="slots-fill-dev"
                style="width: {(devSlots / currentSlots) * 100}%"
              ></div>
            </div>
            <span class="slots-count">{usedSlots}/{currentSlots}</span>
          {:else}
            <span>0</span>
          {/if}
          {#if canManage && isTrial}
            <a
              class="manage-slots-btn"
              href="/{organization}/-/settings/billing"
            >
              Upgrade to Growth Plan
            </a>
          {:else if canManage && !isChcAutoScaled && !isChcHibernated}
            <button
              class="manage-slots-btn"
              on:click={() => (slotsModalOpen = true)}
            >
              Manage Slots
            </button>
          {/if}
        </span>
      </div>
      {#if isChcAutoScaled}
        <div class="info-row pt-0">
          <span class="info-label"></span>
          <span class="text-fg-secondary text-xs">
            Slots reduced to 1 while ClickHouse Cloud is hibernated. They'll be
            restored when the cluster wakes up.
          </span>
        </div>
      {/if}
    {/if}

  </div>
</OverviewCard>

<ManageSlotsModal
  bind:open={slotsModalOpen}
  {organization}
  {project}
  {currentSlots}
  {isRillManaged}
  {isClickHouseCloud}
  required={slotsRequiredMode}
  viewOnly={isTrial}
  detectedMemoryGb={chcDetectedMemoryGb ?? cloudMaxMemory}
/>

{#if isClickHouseCloud && canManage}
  <ClickHouseCloudKeyModal
    bind:open={chcKeyModalOpen}
    {organization}
    {project}
    {connectorHost}
    onDismiss={handleChcDismiss}
    onSave={handleChcKeySaved}
  />
{/if}

{#if isClickHouseCloud && hasCloudApiKey}
  <ClickHouseCloudDetailsModal
    bind:open={chcDetailsModalOpen}
    {organization}
    {project}
    serviceName={cloudServiceName}
    status={cloudStatus}
    provider={cloudProvider}
    region={cloudRegion}
    minMemoryGb={cloudMinMemory}
    maxMemoryGb={cloudMaxMemory}
    replicas={cloudReplicas}
    on:synced={(e) => {
      if (e.detail?.maxMemoryGb) {
        chcDetectedMemoryGb = e.detail.maxMemoryGb;
        sessionStorage.setItem(chcMemoryKey, String(e.detail.maxMemoryGb));
      }
      $proj.refetch();
      $instanceQuery.refetch();
    }}
  />
{/if}

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
  .slots-pill {
    @apply flex h-2.5 w-28 rounded-full bg-gray-200 overflow-hidden;
  }
  .slots-fill-prod {
    @apply h-full bg-primary-500;
  }
  .slots-fill-dev {
    @apply h-full bg-amber-400;
  }
  .slots-count {
    @apply text-sm text-fg-primary font-medium tabular-nums;
  }
  .manage-slots-btn {
    @apply text-xs text-primary-500 bg-transparent border-none cursor-pointer p-0 no-underline;
  }
  .manage-slots-btn:hover {
    @apply text-primary-600;
  }
</style>
