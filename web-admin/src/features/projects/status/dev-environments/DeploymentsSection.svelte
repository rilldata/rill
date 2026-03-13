<script lang="ts">
  import { page } from "$app/stores";
  import {
    V1DeploymentStatus,
    createAdminServiceDeleteDeployment,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetProject,
    type V1Deployment,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    extractBranchFromPath,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import {
    useAllDeployments,
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
  import { Progress } from "@rilldata/web-common/components/progress";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { EyeIcon, PlayIcon, Trash2Icon } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;
  export let project: string;

  const user = createAdminServiceGetCurrentUser();
  const allDeployments = useAllDeployments(organization, project);
  const deleteMutation = createAdminServiceDeleteDeployment();

  $: projectQuery = createAdminServiceGetProject(organization, project);

  $: activeBranch = extractBranchFromPath($page.url.pathname);
  $: currentUserId = $user.data?.user?.id;
  $: rawDeployments = $allDeployments.data?.deployments ?? [];

  // Slot quotas (project-level)
  $: prodSlots = parseInt($projectQuery.data?.project?.prodSlots ?? "0") || 0;
  $: devSlots = parseInt($projectQuery.data?.project?.devSlots ?? "0") || 0;

  // Deduplicate by branch: keep only the most recently updated deployment per branch.
  // This mirrors BranchSelector logic and hides stale/historical deployments.
  $: deduped = (() => {
    const byBranch = new Map<string, V1Deployment>();
    for (const d of rawDeployments) {
      if (
        d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED ||
        d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING ||
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
  $: activeProd = visibleDeployments.filter(
    (d) => d.environment === "prod" && isActiveDeployment(d),
  );
  $: activeDev = visibleDeployments.filter(
    (d) => d.environment !== "prod" && isActiveDeployment(d),
  );
  $: prodSlotsUsed = activeProd.length * prodSlots;
  $: devSlotsUsed = activeDev.length * devSlots;

  function isProd(d: V1Deployment): boolean {
    return d.environment === "prod";
  }

  function isOwnDeployment(ownerUserId: string | undefined): boolean {
    return !!currentUserId && ownerUserId === currentUserId;
  }

  function deploymentSlots(d: V1Deployment): string {
    if (!isActiveDeployment(d)) return "—";
    return String(isProd(d) ? prodSlots : devSlots);
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
  let deletingIds = new Set<string>();
  let deletedIds = new Set<string>();

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

  <div class="slot-cards">
    <div class="slot-card">
      <div class="slot-card-header">
        <span class="slot-card-label">Production</span>
        <span class="slot-card-value">{prodSlotsUsed} / {prodSlots}</span>
      </div>
      <Progress
        value={prodSlotsUsed}
        max={Math.max(prodSlots, 1)}
        class="h-1.5"
      />
      <span class="slot-card-sub">slots</span>
    </div>
    <div class="slot-card">
      <div class="slot-card-header">
        <span class="slot-card-label">Dev branches</span>
        <span class="slot-card-value">
          {devSlotsUsed} / {devSlots * Math.max(activeDev.length, 1)}
        </span>
      </div>
      <Progress
        value={devSlotsUsed}
        max={Math.max(devSlots * Math.max(activeDev.length, 1), 1)}
        class="h-1.5"
      />
      <span class="slot-card-sub">
        slots ({activeDev.length}
        {activeDev.length === 1 ? "branch" : "branches"} &times; {devSlots} slots
        each)
      </span>
    </div>
  </div>

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
        {@const own = isOwnDeployment(deployment.ownerUserId)}
        {@const deleting = deletingIds.has(deployment.id ?? "")}
        {@const id = deployment.id ?? ""}
        <div class="data-row">
          <div class="pl-4 flex items-center gap-2 truncate">
            <span class="font-mono text-xs truncate">
              {deployment.branch || "main"}
            </span>
            {#if prod}
              <span class="prod-badge">production</span>
            {:else if own}
              <span class="own-badge">You</span>
            {/if}
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
                {#if !prod && own}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    href={editUrl(deployment.branch)}
                    on:click={handleNavClick}
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
                  on:click={handleNavClick}
                >
                  <div class="flex items-center">
                    <EyeIcon size="12px" />
                    <span class="ml-2">{prod ? "View" : "Preview"}</span>
                  </div>
                </DropdownMenu.Item>
                {#if !prod}
                  <DropdownMenu.Item
                    class="font-normal flex items-center"
                    disabled={deleting}
                    on:click={() => requestDelete(id, deployment.branch)}
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
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
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

  .slot-cards {
    @apply grid grid-cols-2 gap-4;
  }

  .slot-card {
    @apply flex flex-col gap-y-1.5 border border-border rounded-sm p-4;
  }

  .slot-card-header {
    @apply flex items-center justify-between;
  }

  .slot-card-label {
    @apply text-sm font-semibold text-fg-primary;
  }

  .slot-card-value {
    @apply text-sm font-medium text-fg-primary;
  }

  .slot-card-sub {
    @apply text-xs text-fg-muted;
  }

  .table-wrapper {
    @apply flex flex-col border rounded-sm overflow-hidden;
  }

  .header-row {
    @apply w-full bg-surface-subtle;
    display: grid;
    grid-template-columns:
      minmax(150px, 3fr) minmax(100px, 2fr) minmax(60px, 1fr)
      minmax(120px, 2fr) 56px;
  }

  .data-row {
    @apply w-full py-3 border-t border-gray-200;
    display: grid;
    grid-template-columns:
      minmax(150px, 3fr) minmax(100px, 2fr) minmax(60px, 1fr)
      minmax(120px, 2fr) 56px;
  }

  .prod-badge {
    @apply shrink-0 text-[10px] text-fg-muted;
  }

  .own-badge {
    @apply shrink-0 text-xs bg-primary-50 text-primary-600 px-1.5 py-0.5 rounded;
  }

  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }
</style>
