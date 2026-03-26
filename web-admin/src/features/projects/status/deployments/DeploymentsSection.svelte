<script lang="ts">
  import { page } from "$app/stores";
  import {
    V1DeploymentStatus,
    createAdminServiceDeleteDeployment,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetProject,
    createAdminServiceListDeployments,
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceStartDeployment,
    createAdminServiceStopDeployment,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListDeploymentsQueryKey,
    type V1Deployment,
    type V1GetProjectResponse,
    type V1ListDeploymentsResponse,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    extractBranchFromPath,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import {
    isActiveDeployment,
    invalidateDeployments,
  } from "@rilldata/web-admin/features/edit-session/use-edit-session";
  import {
    getStatusDotClass,
    getStatusLabel,
  } from "@rilldata/web-admin/features/projects/status/display-utils";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { EyeIcon, PlayIcon, StopCircleIcon, Trash2Icon } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

  export let organization: string;
  export let project: string;

  const TRANSIENT_STATUSES = new Set([
    V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING,
    V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING,
    V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
    V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING,
  ]);

  const user = createAdminServiceGetCurrentUser();
  const orgMembers = createAdminServiceListOrganizationMemberUsers(
    organization,
    { pageSize: 1000 },
  );
  // Uses empty params `{}` so the cache key matches BranchSelector's query.
  const allDeployments = createAdminServiceListDeployments(
    organization,
    project,
    {},
    {
      query: {
        refetchInterval: (query) => {
          const deployments = query.state.data?.deployments;
          if (deployments?.some((d) => TRANSIENT_STATUSES.has(d.status!))) {
            return 2000;
          }
          return false;
        },
      },
    },
  );
  const deleteMutation = createAdminServiceDeleteDeployment();
  const startMutation = createAdminServiceStartDeployment();
  const stopMutation = createAdminServiceStopDeployment();

  $: projectQuery = createAdminServiceGetProject(organization, project);

  $: activeBranch = extractBranchFromPath($page.url.pathname);
  $: currentUserId = $user.data?.user?.id;
  $: userNameMap = new Map(
    ($orgMembers.data?.members ?? []).map((m) => [
      m.userId,
      m.userName || m.userEmail || "Unknown",
    ]),
  );
  $: rawDeployments = $allDeployments.data?.deployments ?? [];

  // Slot quotas (project-level); null means the API didn't return a value
  $: prodSlots =
    $projectQuery.data?.project?.prodSlots != null
      ? parseInt($projectQuery.data.project.prodSlots)
      : null;
  $: devSlots =
    $projectQuery.data?.project?.devSlots != null
      ? parseInt($projectQuery.data.project.devSlots)
      : null;

  // Deduplicate by branch: keep only the most recently updated deployment per branch.
  // This mirrors BranchSelector logic and hides stale/historical deployments.
  $: deduped = (() => {
    const byBranch = new Map<string, V1Deployment>();
    for (const d of rawDeployments) {
      if (
        d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED ||
        deletedIds.has(d.id ?? "")
      )
        continue;
      const key = d.branch ?? "";
      const existing = byBranch.get(key);
      if (!existing || (d.updatedOn ?? "") > (existing.updatedOn ?? "")) {
        byBranch.set(key, d);
      }
    }
    return [...byBranch.values()];
  })();

  // Sort: production first, then by updatedOn desc
  $: visibleDeployments = [...deduped].sort((a, b) => {
    const aIsProd = a.environment === "prod";
    const bIsProd = b.environment === "prod";
    if (aIsProd && !bIsProd) return -1;
    if (!aIsProd && bIsProd) return 1;
    return (b.updatedOn ?? "").localeCompare(a.updatedOn ?? "");
  });

  // Slot usage per environment type

  function isProd(d: V1Deployment): boolean {
    return d.environment === "prod";
  }

  function ownerName(d: V1Deployment): string {
    return userNameMap.get(d.ownerUserId ?? "") ?? "—";
  }

  function deploymentSlots(d: V1Deployment): string {
    if (!isActiveDeployment(d)) return "—";
    const slots = isProd(d) ? prodSlots : devSlots;
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

  function previewUrl(branch: string | undefined): string {
    return `/${organization}/${project}${branchPathPrefix(branch)}`;
  }

  function handleNavClick() {
    requestSkipBranchInjection();
  }

  let openDropdownId = "";
  let startingIds = new Set<string>();
  let stoppingIds = new Set<string>();
  let deletingIds = new Set<string>();
  let deletedIds = new Set<string>();

  async function handleStart(deploymentId: string, branch: string | undefined) {
    openDropdownId = "";
    startingIds.add(deploymentId);
    startingIds = startingIds;
    try {
      await $startMutation.mutateAsync({ deploymentId, data: {} });

      // PENDING is in TRANSIENT_STATUSES, so the 2s polling picks
      // up the real status.
      const listKey = getAdminServiceListDeploymentsQueryKey(
        organization,
        project,
        {},
      );
      queryClient.setQueryData<V1ListDeploymentsResponse>(listKey, (old) => {
        if (!old?.deployments) return old;
        return {
          ...old,
          deployments: old.deployments.map((d) =>
            d.id === deploymentId
              ? {
                  ...d,
                  status: V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
                }
              : d,
          ),
        };
      });
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListDeploymentsQueryKey(organization, project),
        refetchType: "none",
      });

      const projectQueryKey = getAdminServiceGetProjectQueryKey(
        organization,
        project,
        branch ? { branch } : undefined,
      );
      queryClient.setQueryData<V1GetProjectResponse>(projectQueryKey, (old) => {
        if (!old?.deployment) return old;
        return {
          ...old,
          deployment: {
            ...old.deployment,
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
          },
        };
      });
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
        refetchType: "none",
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to start deployment: ${getRpcErrorMessage(err as any)}`,
      });
    } finally {
      startingIds.delete(deploymentId);
      startingIds = startingIds;
    }
  }

  async function handleStop(deploymentId: string, branch: string | undefined) {
    openDropdownId = "";
    stoppingIds.add(deploymentId);
    stoppingIds = stoppingIds;
    try {
      await $stopMutation.mutateAsync({ deploymentId, data: {} });

      // Optimistically update ListDeployments. STOPPING is in
      // TRANSIENT_STATUSES, so the 2s polling picks up the real status.
      const listKey = getAdminServiceListDeploymentsQueryKey(
        organization,
        project,
        {},
      );
      queryClient.setQueryData<V1ListDeploymentsResponse>(listKey, (old) => {
        if (!old?.deployments) return old;
        return {
          ...old,
          deployments: old.deployments.map((d) =>
            d.id === deploymentId
              ? {
                  ...d,
                  status: V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING,
                }
              : d,
          ),
        };
      });
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListDeploymentsQueryKey(organization, project),
        refetchType: "none",
      });

      // Optimistically update GetProject so the parent layout
      // transitions to the "Deployment is stopping..." screen
      // instead of waiting for the next poll (which may be minutes
      // away at the RUNNING poll interval).
      const projectQueryKey = getAdminServiceGetProjectQueryKey(
        organization,
        project,
        branch ? { branch } : undefined,
      );
      queryClient.setQueryData<V1GetProjectResponse>(projectQueryKey, (old) => {
        if (!old?.deployment) return old;
        return {
          ...old,
          deployment: {
            ...old.deployment,
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING,
          },
        };
      });
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
        refetchType: "none",
      });
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to stop deployment: ${getRpcErrorMessage(err as any)}`,
      });
    } finally {
      stoppingIds.delete(deploymentId);
      stoppingIds = stoppingIds;
    }
  }

  // Confirmation dialog state
  let deleteConfirmOpen = false;
  let pendingDeleteId = "";
  let pendingDeleteBranch = "";

  function requestDelete(deploymentId: string, branch: string | undefined) {
    pendingDeleteId = deploymentId;
    pendingDeleteBranch = branch ?? "";
    deleteConfirmOpen = true;
  }

  async function handleDelete() {
    const deploymentId = pendingDeleteId;
    const branch = pendingDeleteBranch;
    deleteConfirmOpen = false;

    deletingIds.add(deploymentId);
    deletingIds = deletingIds;
    try {
      await $deleteMutation.mutateAsync({ deploymentId });
      deletedIds.add(deploymentId);
      deletedIds = deletedIds;
      void invalidateDeployments(organization, project);

      if (branch && branch === activeBranch) {
        requestSkipBranchInjection();
        window.location.href = `/${organization}/${project}/-/status/deployments`;
      }
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to delete deployment: ${getRpcErrorMessage(err as any)}`,
      });
    } finally {
      deletingIds.delete(deploymentId);
      deletingIds = deletingIds;
    }
  }
</script>

<section class="flex flex-col gap-y-5">
  <h2 class="text-lg font-medium">Deployments</h2>

  {#if $allDeployments.isLoading}
    <div class="empty-container">
      <DelayedSpinner isLoading={true} size="20px" />
      <span class="text-sm text-fg-secondary">Loading deployments</span>
    </div>
  {:else if $allDeployments.isError}
    <div class="text-red-500 text-sm">
      Error loading deployments: {$allDeployments.error?.message}
    </div>
  {:else if visibleDeployments.length === 0}
    <div class="empty-container">
      <span class="text-fg-secondary font-semibold text-sm">
        No deployments
      </span>
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
          Slots
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm">
          Last updated
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm"></div>
      </div>
      {#each visibleDeployments as deployment (deployment.id)}
        {@const prod = isProd(deployment)}
        {@const starting =
          startingIds.has(deployment.id ?? "") ||
          deployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING}
        {@const stopping =
          stoppingIds.has(deployment.id ?? "") ||
          deployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING}
        {@const deleting =
          deletingIds.has(deployment.id ?? "") ||
          deployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING}
        {@const canStart =
          !prod &&
          deployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED &&
          !starting}
        {@const canStop = !prod && isActiveDeployment(deployment) && !stopping}
        {@const id = deployment.id ?? ""}
        <div class="data-row">
          <div class="pl-4 flex items-center gap-2 truncate">
            <span class="font-mono text-xs truncate">
              {deployment.branch || "main"}
            </span>
            {#if prod}
              <span class="prod-badge">Production</span>
            {/if}
            {#if (deployment.branch ?? "main") === (activeBranch ?? "main")}
              <span class="current-badge">Current</span>
            {/if}
          </div>
          <div
            class="pl-4 flex items-center text-sm text-fg-secondary truncate"
          >
            {ownerName(deployment)}
          </div>
          <div class="pl-4 flex items-center gap-2 text-sm">
            <span
              class="status-dot {getStatusDotClass(
                deployment.status ??
                  V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
              )}"
            ></span>
            {getStatusLabel(
              deployment.status ??
                V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
            )}
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
                {#if !prod && !!currentUserId && deployment.ownerUserId === currentUserId}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    href={editUrl(deployment.branch)}
                    onclick={handleNavClick}
                  >
                    <div class="flex items-center">
                      <PlayIcon size="12px" />
                      <span class="ml-2">Open editor</span>
                    </div>
                  </DropdownMenu.Item>
                {/if}
                <DropdownMenu.Item
                  class="font-normal flex items-center"
                  href={prod
                    ? `/${organization}/${project}`
                    : previewUrl(deployment.branch)}
                  onclick={handleNavClick}
                >
                  <div class="flex items-center">
                    <EyeIcon size="12px" />
                    <span class="ml-2">{prod ? "View" : "Preview"}</span>
                  </div>
                </DropdownMenu.Item>
                {#if canStart}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    onclick={() => handleStart(id, deployment.branch)}
                  >
                    <div class="flex items-center">
                      <PlayIcon size="12px" />
                      <span class="ml-2">Start</span>
                    </div>
                  </DropdownMenu.Item>
                {/if}
                {#if canStop}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    onclick={() => handleStop(id, deployment.branch)}
                  >
                    <div class="flex items-center">
                      <StopCircleIcon size="12px" />
                      <span class="ml-2">Stop</span>
                    </div>
                  </DropdownMenu.Item>
                {/if}
                {#if !prod}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    disabled={deleting}
                    onclick={() => requestDelete(id, deployment.branch)}
                  >
                    <div class="flex items-center">
                      <Trash2Icon size="12px" />
                      <span class="ml-2"
                        >{deleting ? "Deleting..." : "Delete"}</span
                      >
                    </div>
                  </DropdownMenu.Item>
                {/if}
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>

<AlertDialog bind:open={deleteConfirmOpen}>
  <AlertDialogTrigger class="hidden" />
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete this deployment?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          The deployment on branch <span class="font-mono text-xs font-medium"
            >{pendingDeleteBranch || "main"}</span
          > will be deleted. Any unpushed changes will be lost.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          deleteConfirmOpen = false;
        }}
      >
        Cancel
      </Button>
      <Button type="destructive" onClick={handleDelete}>Yes, delete</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>

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
    @apply w-full py-3 border-t border-gray-200;
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
</style>
