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
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    EyeIcon,
    GitBranchIcon,
    PlayIcon,
    StopCircleIcon,
    Trash2Icon,
  } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let orgMembers = $derived(
    createAdminServiceListOrganizationMemberUsers(organization, {
      pageSize: 1000,
    }),
  );
  // Uses empty params `{}` so the cache key matches BranchSelector's query.
  let allDeployments = $derived(
    createAdminServiceListDeployments(
      organization,
      project,
      {},
      {
        query: {
          refetchInterval: (query) => {
            const deployments = query.state.data?.deployments;
            if (deployments?.some((d) => isTransitoryStatus(d.status!))) {
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
  const startMutation = createAdminServiceStartDeployment();
  const stopMutation = createAdminServiceStopDeployment();
  const deleteMutation = createAdminServiceDeleteDeployment();

  let primaryBranch = $derived($projectQuery.data?.project?.primaryBranch);
  let activeBranch = $derived(extractBranchFromPath(page.url.pathname));

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

  let visibleDeployments = $derived.by(() => {
    const active = ($allDeployments.data?.deployments ?? []).filter(
      (d: V1Deployment) =>
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
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

  function previewUrl(branch: string | undefined): string {
    return `/${organization}/${project}${branchPathPrefix(branch)}`;
  }

  let openDropdownId = $state("");
  let pendingId = $state("");
  let deleteDialogOpen = $state(false);
  let pendingDelete = $state<{ id: string; branch: string } | null>(null);

  async function mutateDeployment(
    deploymentId: string,
    branch: string | undefined,
    optimisticStatus: V1DeploymentStatus,
    mutateFn: (args: {
      deploymentId: string;
      data: object;
    }) => Promise<unknown>,
    actionName: string,
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
        message: `Failed to ${actionName} branch: ${getRpcErrorMessage(err as any)}`,
      });
    } finally {
      pendingId = "";
    }
  }

  function confirmDelete(deploymentId: string, branch: string | undefined) {
    pendingDelete = { id: deploymentId, branch: branch ?? "" };
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
        message: `Failed to delete branch: ${getRpcErrorMessage(err as any)}`,
      });
    } finally {
      pendingId = "";
    }
  }
</script>

<section class="flex flex-col gap-y-5">
  <h2 class="text-lg font-medium">Branches</h2>

  {#if $allDeployments.isLoading}
    <div class="empty-container">
      <DelayedSpinner isLoading={true} size="20px" />
      <span class="text-sm text-fg-secondary">Loading branches</span>
    </div>
  {:else if $allDeployments.isError}
    <div class="text-red-500 text-sm">
      Error loading branches: {$allDeployments.error?.message}
    </div>
  {:else if visibleDeployments.length === 0}
    <div class="empty-container">
      <span class="text-fg-secondary font-semibold text-sm"> No branches </span>
    </div>
  {:else}
    <div class="table-wrapper">
      <div class="header-row">
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          Branch
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          Author
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          Status
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          Units
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          Last updated
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
        <div class="data-row">
          <div class="pl-4 flex items-center gap-2 truncate">
            <span class="font-mono text-xs truncate">
              {deployment.branch || primaryBranch || "main"}
            </span>
            {#if prod}
              <span class="prod-badge">Production</span>
            {/if}
            {#if isCurrent}
              <span class="current-badge">Current</span>
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
                <DropdownMenu.Item
                  class="font-normal flex items-center"
                  href={prod
                    ? `/${organization}/${project}`
                    : previewUrl(deployment.branch)}
                  onclick={requestSkipBranchInjection}
                >
                  <div class="flex items-center">
                    <EyeIcon size="12px" />
                    <span class="ml-2">{prod ? "View" : "Preview"}</span>
                  </div>
                </DropdownMenu.Item>
                {#if canStart}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    onclick={() =>
                      mutateDeployment(
                        id,
                        deployment.branch,
                        V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
                        $startMutation.mutateAsync,
                        "resume",
                      )}
                  >
                    <div class="flex items-center">
                      <PlayIcon size="12px" />
                      <span class="ml-2">Resume</span>
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
                        "hibernate",
                      )}
                  >
                    <div class="flex items-center">
                      <StopCircleIcon size="12px" />
                      <span class="ml-2">Hibernate</span>
                    </div>
                  </DropdownMenu.Item>
                {/if}
                {#if !prod}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    disabled={isPending}
                    onclick={() => confirmDelete(id, deployment.branch)}
                  >
                    <div class="flex items-center">
                      <Trash2Icon size="12px" />
                      <span class="ml-2">Delete</span>
                    </div>
                  </DropdownMenu.Item>
                {/if}
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          </div>
        </div>
      {/each}
      <div class="branch-hint">
        <GitBranchIcon size="14" class="shrink-0 text-fg-muted" />
        <span class="text-xs text-fg-secondary">
          Add a branch from the CLI:
        </span>
        <CopyableCodeBlock code="rill project deployment create <branch>" />
      </div>
    </div>
  {/if}
</section>

<DeleteBranchConfirmDialog
  bind:open={deleteDialogOpen}
  branch={pendingDelete?.branch || primaryBranch || "main"}
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

  .current-badge {
    @apply shrink-0 text-xs bg-gray-100 text-fg-muted px-1.5 py-0.5 rounded;
  }

  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }

  .branch-hint {
    @apply flex items-center gap-2 border-t border-border px-4 py-3;
  }
</style>
