<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceGetCurrentUser,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { getStatusDotClass } from "@rilldata/web-admin/features/projects/status/display-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { PlusIcon } from "lucide-svelte";
  import {
    isActiveDeployment,
    useDevDeployments,
    useCreateDevDeployment,
    invalidateDeployments,
  } from "./use-edit-session";

  export let organization: string;
  export let project: string;
  /** The branch currently being viewed (from the URL), if any. */
  export let activeBranch: string | undefined = undefined;

  const user = createAdminServiceGetCurrentUser();
  const devDeployments = useDevDeployments(organization, project);
  const createMutation = useCreateDevDeployment();

  let open = false;
  let branchName = "";
  let showNewBranchInput = false;

  $: currentUserId = $user.data?.user?.id;
  $: deployments = $devDeployments.data?.deployments ?? [];
  $: isLoading = $devDeployments.isLoading;

  // All active deployments owned by the current user, sorted by most recently updated
  $: ownDeployments = deployments
    .filter((d) => d.ownerUserId === currentUserId && isActiveDeployment(d))
    .sort((a, b) => (b.updatedOn ?? "").localeCompare(a.updatedOn ?? ""));

  // If viewing a branch the user owns, clicking the button should go straight there
  $: activeBranchDeployment = activeBranch
    ? ownDeployments.find((d) => d.branch === activeBranch)
    : undefined;

  $: hasOwnSessions = ownDeployments.length > 0;
  $: isStarting = $createMutation.isPending;

  // Reset state when dropdown opens
  $: if (open) {
    branchName = "";
    showNewBranchInput = !hasOwnSessions;
  }

  function editUrl(branch: string | undefined): string {
    return `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  }

  function statusDot(status: V1DeploymentStatus | undefined): string {
    return getStatusDotClass(
      status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
    );
  }

  // When the user owns a deployment on the active branch, the button
  // links directly to that editor (no dropdown).
  $: directEditHref = activeBranchDeployment
    ? editUrl(activeBranchDeployment.branch)
    : undefined;

  function handleButtonClick(e: MouseEvent) {
    e.preventDefault();
    requestSkipBranchInjection();
    void goto(directEditHref!);
  }

  function handleNavClick() {
    requestSkipBranchInjection();
  }

  async function handleCreate() {
    if (!branchName.trim()) return;
    try {
      const resp = await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
          branch: branchName.trim(),
        },
      });
      void invalidateDeployments(organization, project);
      open = false;
      requestSkipBranchInjection();
      await goto(editUrl(resp.deployment?.branch));
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to start edit session: ${getRpcErrorMessage(err as any)}`,
      });
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      void handleCreate();
    }
  }
</script>

{#if directEditHref}
  <!-- On a branch the user owns: navigate directly, no dropdown -->
  <Button
    type="secondary"
    href={directEditHref}
    disabled={isStarting || isLoading}
    onClick={handleButtonClick}
  >
    Edit
  </Button>
{:else}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button
          {...props}
          type="secondary"
          disabled={isStarting || isLoading}
          loading={isStarting}
          loadingCopy="Starting..."
        >
          Edit
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>

    <DropdownMenu.Content align="end" class="min-w-[200px] max-w-[280px]">
      {#if hasOwnSessions}
        <DropdownMenu.Group>
          <DropdownMenu.Label>Your branches</DropdownMenu.Label>
        </DropdownMenu.Group>
        {#each ownDeployments as deployment (deployment.id)}
          <DropdownMenu.Item
            href={editUrl(deployment.branch)}
            onclick={handleNavClick}
          >
            <span
              class="inline-block size-1.5 rounded-full flex-none {statusDot(
                deployment.status,
              )}"
            ></span>
            <span class="font-mono truncate">
              {deployment.branch || "main"}
            </span>
          </DropdownMenu.Item>
        {/each}

        <DropdownMenu.Separator />

        {#if showNewBranchInput}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="flex flex-col gap-y-1.5 px-2 pb-1.5 pt-0.5"
            onclick={(e) => e.stopPropagation()}
          >
            <!-- svelte-ignore a11y_autofocus -->
            <input
              class="branch-input"
              type="text"
              bind:value={branchName}
              onkeydown={(e) => {
                e.stopPropagation();
                handleKeydown(e);
              }}
              placeholder="branch-name"
              autofocus
            />
            <Button
              type="primary"
              small
              disabled={!branchName.trim() || isStarting}
              loading={isStarting}
              loadingCopy="Starting..."
              onClick={handleCreate}
            >
              Start editing
            </Button>
          </div>
        {:else}
          <!-- Raw button (not DropdownMenu.Item) so clicking doesn't close the menu -->
          <button
            class="new-branch-btn"
            onclick={(e) => {
              e.stopPropagation();
              showNewBranchInput = true;
            }}
          >
            <PlusIcon size="12" />
            <span>New branch...</span>
          </button>
        {/if}
      {:else}
        <DropdownMenu.Group>
          <DropdownMenu.Label>Create a branch</DropdownMenu.Label>
        </DropdownMenu.Group>
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          class="flex flex-col gap-y-1.5 px-2 pb-1.5 pt-0.5"
          onclick={(e) => e.stopPropagation()}
        >
          <!-- svelte-ignore a11y_autofocus -->
          <input
            class="branch-input"
            type="text"
            bind:value={branchName}
            onkeydown={(e) => {
              e.stopPropagation();
              handleKeydown(e);
            }}
            placeholder="branch-name"
            autofocus
          />
          <Button
            type="primary"
            small
            disabled={!branchName.trim() || isStarting}
            loading={isStarting}
            loadingCopy="Starting..."
            onClick={handleCreate}
          >
            Start editing
          </Button>
        </div>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<style lang="postcss">
  .new-branch-btn {
    @apply flex w-full items-center gap-x-2 rounded-sm px-2 py-1.5 text-xs;
    @apply text-primary-600 hover:bg-surface-hover cursor-pointer;
  }

  .branch-input {
    @apply w-full text-xs font-mono px-2 py-1 rounded border border-gray-300;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
  }
</style>
