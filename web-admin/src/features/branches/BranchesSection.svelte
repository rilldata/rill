<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import {
    V1DeploymentStatus,
    createAdminServiceDeleteDeployment,
    createAdminServiceGetProject,
    createAdminServiceListDeployments,
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceStartDeployment,
    createAdminServiceStopDeployment,
    type V1Deployment,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    extractBranchFromPath,
    requestSkipBranchInjection,
  } from "./branch-utils";
  import { isActiveDeployment, isProdDeployment } from "./deployment-utils";
  import {
    getStatusDotClass,
    getStatusLabel,
    isTransitoryStatus,
  } from "@rilldata/web-admin/features/projects/status/display-utils";
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
  import {
    optimisticallyRemoveDeployment,
    optimisticallySetStatus,
  } from "./branch-actions";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import DeleteBranchConfirmDialog from "./DeleteBranchConfirmDialog.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CopyableCodeBlock from "@rilldata/web-common/components/calls-to-action/CopyableCodeBlock.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type { FilterGroup } from "@rilldata/web-common/components/table-toolbar/types";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import {
    GitBranchIcon,
    PlayIcon,
    StopCircleIcon,
    Trash2Icon,
  } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { onMount } from "svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let orgMembers = $derived(
    createAdminServiceListOrganizationMemberUsers(organization, {
      pageSize: 1000,
    }),
  );
  // Uses empty params `{}` so the cache key matches other unscoped consumers.
  let allDeployments = $derived(
    createAdminServiceListDeployments(
      organization,
      project,
      {},
      {
        query: {
          refetchInterval: (query) => {
            const deployments = query.state.data?.deployments;
            if (
              deployments?.some((d) =>
                isTransitoryStatus(
                  d.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
                ),
              )
            ) {
              return 2000;
            }
            return false;
          },
        },
      },
    ),
  );
  let projectQuery = $derived(
    createAdminServiceGetProject(organization, project),
  );
  const { cloudEditing } = featureFlags;
  const startMutation = createAdminServiceStartDeployment();
  const stopMutation = createAdminServiceStopDeployment();
  const deleteMutation = createAdminServiceDeleteDeployment();

  let primaryBranch = $derived($projectQuery.data?.project?.primaryBranch);
  let activeBranch = $derived(extractBranchFromPath(page.url.pathname));
  let canEditProject = $derived(
    $cloudEditing && !!$projectQuery.data?.projectPermissions?.manageDev,
  );

  let userNameMap = $derived(
    new Map(
      ($orgMembers.data?.members ?? []).map((m) => [
        m.userId,
        m.userName || m.userEmail || "Unknown",
      ]),
    ),
  );

  let prodSlots = $derived(
    $projectQuery.data?.project?.prodSlots != null
      ? parseInt($projectQuery.data.project.prodSlots, 10)
      : null,
  );
  let devSlots = $derived(
    $projectQuery.data?.project?.devSlots != null
      ? parseInt($projectQuery.data.project.devSlots, 10)
      : null,
  );

  // Toolbar state — synced to URL params `q` and `status` (multi-select array)
  const filterSync = createUrlFilterSync([
    { key: "q", type: "string" },
    { key: "status", type: "array" },
  ]);

  let searchText = $state(parseStringParam(page.url.searchParams.get("q")));
  let statusFilter = $state<string[]>(
    parseArrayParam(page.url.searchParams.get("status")),
  );
  let mounted = $state(false);

  onMount(() => {
    filterSync.init(page.url);
    mounted = true;
  });

  // URL → local state on external navigation (back/forward)
  $effect(() => {
    if (!mounted) return;
    const url = page.url;
    if (filterSync.hasExternalNavigation(url)) {
      filterSync.markSynced(url);
      searchText = parseStringParam(url.searchParams.get("q"));
      statusFilter = parseArrayParam(url.searchParams.get("status"));
    }
  });

  // Local state → URL
  $effect(() => {
    if (!mounted) return;
    filterSync.syncToUrl({ q: searchText, status: statusFilter });
  });

  let filterGroups = $derived([
    {
      label: m.common_status(),
      key: "status",
      options: [
        { label: m.branch_status_ready(), value: "running" },
        { label: m.branch_status_pending(), value: "pending" },
        { label: m.branch_status_error(), value: "errored" },
        { label: m.branch_status_stopped(), value: "stopped" },
      ],
      selected: statusFilter,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[]);

  function statusMatches(d: V1Deployment): boolean {
    if (statusFilter.length === 0) return true;
    const s = d.status;
    return statusFilter.some((sel) => {
      switch (sel) {
        case "running":
          return s === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING;
        case "pending":
          return (
            s === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
            s === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING
          );
        case "errored":
          return s === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED;
        case "stopped":
          return (
            s === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED ||
            s === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING
          );
        default:
          return false;
      }
    });
  }

  let visibleDeployments = $derived.by(() => {
    const q = searchText.trim().toLowerCase();
    const active = ($allDeployments.data?.deployments ?? []).filter(
      (d: V1Deployment) =>
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED &&
        statusMatches(d) &&
        (q === "" || (d.branch ?? "").toLowerCase().includes(q)),
    );
    return [...active].sort((a, b) => {
      const aIsProd = isProdDeployment(a);
      const bIsProd = isProdDeployment(b);
      if (aIsProd && !bIsProd) return -1;
      if (!aIsProd && bIsProd) return 1;
      return (a.branch ?? "").localeCompare(b.branch ?? "");
    });
  });

  function ownerName(d: V1Deployment): string {
    return userNameMap.get(d.ownerUserId ?? "") ?? "—";
  }

  function deploymentSlots(d: V1Deployment): string {
    if (!isActiveDeployment(d)) return "—";
    const slots = isProdDeployment(d) ? prodSlots : devSlots;
    return slots != null ? String(slots) : "—";
  }

  function formatDate(dateStr: string | undefined): string {
    if (!dateStr) return "—";
    return new Date(dateStr).toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }

  function editUrl(branch: string | undefined): string {
    return `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  }

  let openDropdownId = $state("");
  let pendingId = $state("");
  let deleteDialogOpen = $state(false);
  let pendingDelete = $state<{
    id: string;
    branch: string;
    editable: boolean;
  } | null>(null);

  function onFilterChange(key: string, selected: string[]) {
    if (key === "status") statusFilter = selected;
  }

  async function mutateDeployment(
    deploymentId: string,
    branch: string | undefined,
    optimisticStatus: V1DeploymentStatus,
    mutateFn: (args: {
      deploymentId: string;
      data: object;
    }) => Promise<unknown>,
    errorMsg: (error: string) => string,
  ) {
    openDropdownId = "";
    pendingId = deploymentId;
    try {
      await mutateFn({ deploymentId, data: {} });
      optimisticallySetStatus(
        organization,
        project,
        deploymentId,
        branch,
        optimisticStatus,
      );
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: errorMsg(getRpcErrorMessage(err)),
      });
    } finally {
      pendingId = "";
    }
  }

  function confirmDelete(
    deploymentId: string,
    branch: string | undefined,
    editable: boolean,
  ) {
    pendingDelete = { id: deploymentId, branch: branch ?? "", editable };
    deleteDialogOpen = true;
  }

  async function handleDelete() {
    if (!pendingDelete) return;
    const { id: deploymentId, branch } = pendingDelete;
    pendingDelete = null;
    pendingId = deploymentId;
    try {
      await $deleteMutation.mutateAsync({ deploymentId });
      optimisticallyRemoveDeployment(organization, project, deploymentId);
      if (branch && branch === activeBranch) {
        requestSkipBranchInjection();
        void goto(`/${organization}/${project}/-/status/branches`);
      }
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: m.branch_delete_failed({ error: getRpcErrorMessage(err) }),
      });
    } finally {
      pendingId = "";
    }
  }
</script>

<section class="flex flex-col gap-y-5">
  <h2 class="text-lg font-medium">{m.branch_branches()}</h2>

  <TableToolbar
    bind:searchText
    {filterGroups}
    {onFilterChange}
    onClearAllFilters={() => {
      statusFilter = [];
      searchText = "";
    }}
    showSort={false}
  />

  {#if $allDeployments.isLoading}
    <div class="empty-container">
      <DelayedSpinner isLoading={true} size="20px" />
      <span class="text-sm text-fg-secondary">{m.branch_loading()}</span>
    </div>
  {:else if $allDeployments.isError}
    <div class="text-red-500 text-sm">
      {m.branch_error_loading({ error: $allDeployments.error?.message })}
    </div>
  {:else if visibleDeployments.length === 0}
    <div class="empty-container">
      <span class="text-fg-secondary font-semibold text-sm">{m.branch_no_branches()}</span>
    </div>
  {:else}
    <div class="table-wrapper">
      <div class="header-row">
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          {m.branch_branch()}
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          {m.branch_author()}
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          {m.common_status()}
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          {m.branch_slots()}
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          {m.branch_last_updated()}
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm"></div>
      </div>
      {#each visibleDeployments as deployment, i (deployment.id ?? i)}
        {@const prod = isProdDeployment(deployment)}
        {@const id = deployment.id ?? ""}
        {@const status =
          deployment.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED}
        {@const isPending = id === pendingId}
        {@const isCurrent = prod
          ? !activeBranch
          : activeBranch === deployment.branch}
        {@const canStart =
          !prod &&
          status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED &&
          !isPending}
        {@const canStop = !prod && isActiveDeployment(deployment) && !isPending}
        {@const branchName = deployment.branch || primaryBranch || "main"}
        <div class="data-row">
          <div class="pl-4 flex items-center gap-2 truncate">
            <span class="font-mono text-xs truncate" title={branchName}>
              {branchName}
            </span>
            {#if prod}
              <span class="prod-badge">{m.branch_production()}</span>
            {/if}
            {#if !prod && !deployment.editable}
              <Tooltip location="bottom" distance={8}>
                <span class="readonly-badge">{m.branch_read_only()}</span>
                <TooltipContent slot="tooltip-content">
                  <div class="text-xs max-w-[360px] flex flex-col gap-y-1">
                    <span>{m.branch_not_editable()}</span>
                    <span>
                      {m.branch_recreate_with()}
                      <code class="font-mono"
                        >rill project deployment create {branchName} --editable</code
                      >.
                    </span>
                  </div>
                </TooltipContent>
              </Tooltip>
            {/if}
            {#if isCurrent}
              <span class="current-badge">{m.branch_current()}</span>
            {/if}
          </div>
          <div
            class="pl-4 flex items-center text-sm text-fg-secondary truncate"
          >
            {ownerName(deployment)}
          </div>
          <div class="pl-4 flex items-center gap-2 text-sm">
            {#if isTransitoryStatus(status)}
              <LoadingCircleOutline size="12px" />
            {:else}
              <span class="status-dot {getStatusDotClass(status)}"></span>
            {/if}
            {getStatusLabel(status)}
          </div>
          <div class="pl-4 flex items-center text-sm text-fg-secondary">
            {deploymentSlots(deployment)}
          </div>
          <div class="pl-4 flex items-center text-sm text-fg-secondary">
            {formatDate(deployment.updatedOn)}
          </div>
          <div class="pl-4 flex items-center">
            {#if !prod}
              <DropdownMenu.Root
                open={openDropdownId === id}
                onOpenChange={(open) => {
                  openDropdownId = open ? id : "";
                }}
              >
                <DropdownMenu.Trigger class="flex-none">
                  <IconButton rounded active={openDropdownId === id} size={20}>
                    <ThreeDot size="16px" />
                  </IconButton>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="start">
                  {#if canEditProject && deployment.editable}
                    <DropdownMenu.Item
                      class="font-normal flex items-center"
                      href={editUrl(deployment.branch)}
                      onclick={requestSkipBranchInjection}
                    >
                      <div class="flex items-center">
                        <PlayIcon size="12px" />
                        <span class="ml-2">{m.branch_open_editor()}</span>
                      </div>
                    </DropdownMenu.Item>
                  {/if}
                  {#if canStart}
                    <DropdownMenu.Item
                      class="font-normal flex items-center"
                      onclick={() =>
                        mutateDeployment(
                          id,
                          deployment.branch,
                          V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
                          $startMutation.mutateAsync,
                          (error) => m.branch_resume_failed({ error }),
                        )}
                    >
                      <div class="flex items-center">
                        <PlayIcon size="12px" />
                        <span class="ml-2">{m.branch_resume()}</span>
                      </div>
                    </DropdownMenu.Item>
                  {/if}
                  {#if canStop}
                    <DropdownMenu.Item
                      class="font-normal flex items-center"
                      onclick={() =>
                        mutateDeployment(
                          id,
                          deployment.branch,
                          V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING,
                          $stopMutation.mutateAsync,
                          (error) => m.branch_hibernate_failed({ error }),
                        )}
                    >
                      <div class="flex items-center">
                        <StopCircleIcon size="12px" />
                        <span class="ml-2">{m.branch_hibernate()}</span>
                      </div>
                    </DropdownMenu.Item>
                  {/if}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    disabled={isPending}
                    onclick={() =>
                      confirmDelete(
                        id,
                        deployment.branch,
                        !!deployment.editable,
                      )}
                  >
                    <div class="flex items-center">
                      <Trash2Icon size="12px" />
                      <span class="ml-2">{m.common_delete()}</span>
                    </div>
                  </DropdownMenu.Item>
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            {/if}
          </div>
        </div>
      {/each}
      <div class="branch-hint">
        <GitBranchIcon size="14" class="shrink-0 text-fg-muted" />
        <span class="text-xs text-fg-secondary">
          {m.branch_add_from_cli()}
        </span>
        <CopyableCodeBlock code="rill project deployment create <branch>" />
      </div>
    </div>
  {/if}
</section>

<DeleteBranchConfirmDialog
  bind:open={deleteDialogOpen}
  branch={pendingDelete?.branch || primaryBranch || "main"}
  editable={pendingDelete?.editable ?? false}
  onConfirm={handleDelete}
/>

<style lang="postcss">
  .empty-container {
    @apply border border-border rounded-sm py-10 flex flex-col items-center gap-y-2;
  }

  .table-wrapper {
    @apply flex flex-col border rounded-sm overflow-x-auto;
  }

  .header-row,
  .data-row {
    display: grid;
    grid-template-columns:
      minmax(150px, 3fr) minmax(100px, 2fr) minmax(100px, 2fr)
      minmax(60px, 1fr) minmax(120px, 2fr) 56px;
    min-width: 686px;
  }

  .header-row {
    @apply w-full bg-surface-subtle;
  }

  .data-row {
    @apply w-full py-3 border-t border-border;
  }

  .prod-badge {
    @apply shrink-0 text-xs bg-primary-50 text-primary-600 px-1.5 py-0.5 rounded;
  }

  .current-badge,
  .readonly-badge {
    @apply shrink-0 text-xs bg-gray-100 text-fg-muted px-1.5 py-0.5 rounded;
  }

  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }

  .branch-hint {
    @apply flex items-center gap-2 border-t border-border px-4 py-3;
  }
</style>
