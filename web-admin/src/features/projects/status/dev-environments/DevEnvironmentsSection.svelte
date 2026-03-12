<script lang="ts">
  import {
    V1DeploymentStatus,
    createAdminServiceDeleteDeployment,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import {
    useDevDeployments,
    invalidateDeployments,
  } from "@rilldata/web-admin/features/edit-session/use-edit-session";
  import {
    getStatusDotClass,
    getStatusLabel,
  } from "@rilldata/web-admin/features/projects/status/display-utils";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { EyeIcon, PlayIcon, Trash2Icon } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let organization: string;
  export let project: string;

  const user = createAdminServiceGetCurrentUser();
  const devDeployments = useDevDeployments(organization, project);
  const deleteMutation = createAdminServiceDeleteDeployment();

  $: currentUserId = $user.data?.user?.id;
  $: deployments = $devDeployments.data?.deployments ?? [];
  $: visibleDeployments = deployments
    .filter(
      (d) =>
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED &&
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING &&
        !deletedIds.has(d.id ?? ""),
    )
    .sort((a, b) => (b.updatedOn ?? "").localeCompare(a.updatedOn ?? ""));

  function isOwnDeployment(ownerUserId: string | undefined): boolean {
    return !!currentUserId && ownerUserId === currentUserId;
  }

  function isActive(status: V1DeploymentStatus | undefined): boolean {
    return (
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
      status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING
    );
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
  async function handleDelete(deploymentId: string) {
    deletingIds.add(deploymentId);
    deletingIds = deletingIds;
    try {
      await $deleteMutation.mutateAsync({ deploymentId });
      // Optimistically hide the row; the server-side status transition is async
      deletedIds.add(deploymentId);
      deletedIds = deletedIds;
      void invalidateDeployments(organization, project);
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to delete environment: ${getRpcErrorMessage(err as any)}`,
      });
    } finally {
      deletingIds.delete(deploymentId);
      deletingIds = deletingIds;
    }
  }
</script>

<section class="flex flex-col gap-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Dev Environments</h2>
  </div>

  {#if $devDeployments.isLoading}
    <div class="empty-container">
      <DelayedSpinner isLoading={true} size="20px" />
      <span class="text-sm text-fg-secondary">Loading environments</span>
    </div>
  {:else if $devDeployments.isError}
    <div class="text-red-500 text-sm">
      Error loading environments: {$devDeployments.error?.message}
    </div>
  {:else if visibleDeployments.length === 0}
    <div class="empty-container">
      <span class="text-fg-secondary font-semibold text-sm">
        No dev environments
      </span>
      <span class="text-fg-muted text-sm">
        Use the "Edit" button in the header to start editing this project.
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
          Last updated
        </div>
        <div class="pl-4 py-2 font-semibold text-fg-secondary text-sm"></div>
      </div>
      {#each visibleDeployments as deployment (deployment.id)}
        {@const own = isOwnDeployment(deployment.ownerUserId)}
        {@const active = isActive(deployment.status)}
        {@const deleting = deletingIds.has(deployment.id ?? "")}
        {@const id = deployment.id ?? ""}
        <div class="data-row">
          <div class="pl-4 flex items-center gap-2 truncate">
            <span class="font-mono text-xs truncate">
              {deployment.branch || "main"}
            </span>
            {#if own}
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
                {#if own}
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
                  href={previewUrl(deployment.branch)}
                  on:click={handleNavClick}
                >
                  <div class="flex items-center">
                    <EyeIcon size="12px" />
                    <span class="ml-2">Preview</span>
                  </div>
                </DropdownMenu.Item>
                <DropdownMenu.Item
                  class="font-normal flex items-center"
                  disabled={deleting}
                  on:click={() => handleDelete(id)}
                >
                  <div class="flex items-center">
                    <Trash2Icon size="12px" />
                    <span class="ml-2"
                      >{deleting ? "Deleting..." : "Delete"}</span
                    >
                  </div>
                </DropdownMenu.Item>
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>

<style lang="postcss">
  .empty-container {
    @apply border border-border rounded-sm py-10 flex flex-col items-center gap-y-2;
  }

  .table-wrapper {
    @apply flex flex-col border rounded-sm overflow-hidden;
  }

  .header-row {
    @apply w-full bg-surface-subtle;
    display: grid;
    grid-template-columns:
      minmax(150px, 3fr) minmax(100px, 2fr) minmax(120px, 2fr)
      56px;
  }

  .data-row {
    @apply w-full py-3 border-t border-gray-200;
    display: grid;
    grid-template-columns:
      minmax(150px, 3fr) minmax(100px, 2fr) minmax(120px, 2fr)
      56px;
  }

  .own-badge {
    @apply shrink-0 text-xs bg-primary-50 text-primary-600 px-1.5 py-0.5 rounded;
  }

  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }
</style>
