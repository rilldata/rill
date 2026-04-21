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
  import * as Popover from "@rilldata/web-common/components/popover";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { PlusIcon } from "lucide-svelte";
  import {
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

  // Editable deployments owned by the current user (excludes ones being
  // torn down), sorted by most recently updated. Stopped and errored
  // branches are shown so the user can resume or retry them. Non-editable
  // deployments (e.g. created via the CLI without `--editable`) are hidden
  // because the edit surface cannot function against them.
  $: ownDeployments = deployments
    .filter(
      (d) =>
        d.ownerUserId === currentUserId &&
        d.editable &&
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING &&
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
    )
    .sort((a, b) => (b.updatedOn ?? "").localeCompare(a.updatedOn ?? ""));

  // If viewing a branch the user owns, clicking the button should go straight there
  $: activeBranchDeployment = activeBranch
    ? ownDeployments.find((d) => d.branch === activeBranch)
    : undefined;

  // True when the active branch has a deployment the user owns but which
  // isn't editable (e.g. created via the CLI without `--editable`). Used to
  // show a dropdown banner explaining that the user needs a new branch.
  $: activeBranchIsNonEditable =
    !!activeBranch &&
    !!currentUserId &&
    deployments.some(
      (d) =>
        d.branch === activeBranch &&
        d.ownerUserId === currentUserId &&
        !d.editable,
    );

  $: hasOwnSessions = ownDeployments.length > 0;
  $: isStarting = $createMutation.isPending;

  // Reset state when popover opens
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
  // links directly to that editor (no popover).
  $: directEditHref = activeBranchDeployment
    ? editUrl(activeBranchDeployment.branch)
    : undefined;

  function handleButtonClick(e: MouseEvent) {
    e.preventDefault();
    requestSkipBranchInjection();
    void goto(directEditHref!);
  }

  function handleBranchClick() {
    requestSkipBranchInjection();
    open = false;
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
  <!-- On a branch the user owns: navigate directly, no popover -->
  <Button
    type="secondary"
    href={directEditHref}
    disabled={isStarting || isLoading}
    onClick={handleButtonClick}
  >
    Edit
  </Button>
{:else}
  <Popover.Root bind:open>
    <Popover.Trigger>
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
    </Popover.Trigger>

    <Popover.Content
      align="end"
      padding="1.5"
      class="w-auto min-w-[200px] max-w-[280px]"
    >
      {#if activeBranchIsNonEditable}
        <div class="banner">
          This branch isn't editable. Start a new one below{hasOwnSessions
            ? " or switch to another"
            : ""}.
        </div>
      {/if}
      {#if hasOwnSessions}
        <div class="section-label">Your branches</div>
        {#each ownDeployments as deployment (deployment.id)}
          <a
            class="branch-row"
            href={editUrl(deployment.branch)}
            onclick={handleBranchClick}
            data-sveltekit-preload-data="hover"
          >
            <span
              class="inline-block size-1.5 rounded-full flex-none {statusDot(
                deployment.status,
              )}"
            ></span>
            <span class="font-mono truncate">
              {deployment.branch || "main"}
            </span>
          </a>
        {/each}

        <div class="separator"></div>

        {#if showNewBranchInput}
          <div class="flex flex-col gap-y-1.5 px-2 pb-1.5 pt-0.5">
            <!-- svelte-ignore a11y_autofocus -->
            <input
              class="branch-input"
              type="text"
              bind:value={branchName}
              onkeydown={handleKeydown}
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
          <button
            class="new-branch-btn"
            onclick={() => (showNewBranchInput = true)}
          >
            <PlusIcon size="12" />
            <span>New branch...</span>
          </button>
        {/if}
      {:else}
        <div class="section-label">Create a branch</div>
        <div class="flex flex-col gap-y-1.5 px-2 pb-1.5 pt-0.5">
          <!-- svelte-ignore a11y_autofocus -->
          <input
            class="branch-input"
            type="text"
            bind:value={branchName}
            onkeydown={handleKeydown}
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
    </Popover.Content>
  </Popover.Root>
{/if}

<style lang="postcss">
  .section-label {
    @apply px-2 py-1.5 text-xs text-fg-secondary font-semibold;
  }

  .banner {
    @apply mx-0.5 mt-0.5 mb-1.5 rounded-sm px-2 py-1.5;
    @apply text-xs text-yellow-800 bg-yellow-50 border border-yellow-200;
  }

  .branch-row {
    @apply flex items-center gap-x-2 rounded-sm px-2 py-1.5 text-xs;
    @apply text-fg-primary hover:bg-surface-hover hover:text-fg-accent;
    @apply cursor-pointer outline-none;
  }

  .separator {
    @apply -mx-1 my-1 h-px bg-border;
  }

  .new-branch-btn {
    @apply flex w-full items-center gap-x-2 rounded-sm px-2 py-1.5 text-xs;
    @apply text-primary-600 hover:bg-surface-hover cursor-pointer;
  }

  .branch-input {
    @apply w-full text-xs font-mono px-2 py-1 rounded border border-gray-300;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
  }
</style>
