<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    V1DeploymentStatus,
    createAdminServiceDeleteDeployment,
    createAdminServiceGetCurrentUser,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import {
    useDevDeployments,
    useCreateDevDeployment,
    invalidateDeployments,
  } from "@rilldata/web-admin/features/edit-session/use-edit-session";
  import {
    getStatusDotClass,
    getStatusLabel,
  } from "@rilldata/web-admin/features/projects/status/display-utils";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { PlayIcon, Trash2Icon } from "lucide-svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let organization: string;
  export let project: string;

  const user = createAdminServiceGetCurrentUser();
  const devDeployments = useDevDeployments(organization, project);
  const createMutation = useCreateDevDeployment();
  const deleteMutation = createAdminServiceDeleteDeployment();

  $: currentUserId = $user.data?.user?.id;
  $: deployments = $devDeployments.data?.deployments ?? [];
  $: visibleDeployments = deployments.filter(
    (d) =>
      d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED &&
      d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING,
  );

  $: isCreating = $createMutation.isPending;

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

  async function handleNewSession() {
    try {
      const resp = await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
        },
      });
      void invalidateDeployments(organization, project);
      requestSkipBranchInjection();
      await goto(editUrl(resp.deployment?.branch));
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to start edit session: ${getRpcErrorMessage(err as any)}`,
      });
    }
  }

  async function handleResume(branch: string | undefined) {
    requestSkipBranchInjection();
    await goto(editUrl(branch));
  }

  async function handlePreview(branch: string | undefined) {
    requestSkipBranchInjection();
    await goto(previewUrl(branch));
  }

  let openDropdownId = "";
  let deletingIds = new Set<string>();
  async function handleDelete(deploymentId: string) {
    deletingIds.add(deploymentId);
    deletingIds = deletingIds;
    try {
      await $deleteMutation.mutateAsync({ deploymentId });
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListDeploymentsQueryKey(
          organization,
          project,
          { environment: "dev" },
        ),
      });
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
    <Button
      type="secondary"
      large
      disabled={isCreating}
      loading={isCreating}
      loadingCopy="Starting..."
      onClick={handleNewSession}
    >
      New edit session
    </Button>
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
        Click "New edit session" to start editing this project in the cloud.
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
                    on:click={() => handleResume(deployment.branch)}
                  >
                    <div class="flex items-center">
                      <PlayIcon size="12px" />
                      <span class="ml-2">Resume</span>
                    </div>
                  </DropdownMenu.Item>
                {/if}
                <DropdownMenu.Item
                  class="font-normal flex items-center"
                  on:click={() => handlePreview(deployment.branch)}
                >
                  <div class="flex items-center">
                    <span class="ml-0.5">Preview</span>
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
