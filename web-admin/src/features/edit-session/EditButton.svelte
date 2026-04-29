<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceCreateDeployment,
    createAdminServiceGetCurrentUser,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    branchPathPrefix,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { GitBranchIcon, PlusIcon } from "lucide-svelte";
  import { useDevDeployments, invalidateDeployments } from "./use-edit-session";

  export let organization: string;
  export let project: string;
  /** The branch currently being viewed (from the URL), if any. */
  export let activeBranch: string | undefined = undefined;
  /** The project's primary branch, used as the source for new branches. */
  export let primaryBranch: string | undefined = undefined;

  const user = createAdminServiceGetCurrentUser();
  const devDeployments = useDevDeployments(organization, project);
  const createMutation = createAdminServiceCreateDeployment();

  let open = false;
  let branchName = "";
  let showNewBranchInput = false;

  $: currentUserId = $user.data?.user?.id;
  $: deployments = $devDeployments.data?.deployments ?? [];
  $: isLoading = $devDeployments.isLoading;
  $: sourceBranch = primaryBranch || "main";

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
  // show a banner explaining that the user needs a new branch.
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

  // Reset state when dialog opens
  $: if (open) {
    branchName = "";
    showNewBranchInput = !hasOwnSessions;
  }

  function editUrl(branch: string | undefined): string {
    return `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  }

  // When the user owns a deployment on the active branch, the button
  // links directly to that editor (no dialog).
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

  // Replaces whitespace with "-" as the user types so branch names are
  // always valid. Space → "-" is a 1:1 swap, so cursor stays put.
  function handleBranchNameInput(e: Event) {
    const target = e.currentTarget as HTMLInputElement;
    const sanitized = target.value.replace(/\s+/g, "-");
    if (sanitized !== target.value) {
      const cursorPos = target.selectionStart ?? sanitized.length;
      target.value = sanitized;
      target.setSelectionRange(cursorPos, cursorPos);
    }
    branchName = sanitized;
  }

  function handleCancelNewBranch() {
    branchName = "";
    showNewBranchInput = false;
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
    } else if (e.key === "Escape" && hasOwnSessions && showNewBranchInput) {
      e.preventDefault();
      e.stopPropagation();
      handleCancelNewBranch();
    }
  }
</script>

{#if directEditHref}
  <!-- On a branch the user owns: navigate directly, no dialog -->
  <Button
    type="secondary"
    href={directEditHref}
    disabled={isStarting || isLoading}
    onClick={handleButtonClick}
  >
    Edit
  </Button>
{:else}
  <Dialog.Root bind:open>
    <Dialog.Trigger>
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
    </Dialog.Trigger>

    <Dialog.Content class="max-w-md gap-0">
      <Dialog.Header class="space-y-1">
        <Dialog.Title class="text-base">
          {hasOwnSessions ? "Continue editing" : "Start editing"}
        </Dialog.Title>
        <Dialog.Description class="dlg-subtitle">
          {hasOwnSessions
            ? "Branched from"
            : "Create a branch to edit from"}<span class="branch-chip">
            <GitBranchIcon size="11" />
            {sourceBranch}
          </span>
        </Dialog.Description>
      </Dialog.Header>

      {#if activeBranchIsNonEditable}
        <div class="banner">
          This branch isn't editable. Start a new one below{hasOwnSessions
            ? " or switch to another"
            : ""}.
        </div>
      {/if}

      {#if hasOwnSessions}
        <div class="branch-list">
          {#each ownDeployments as deployment (deployment.id)}
            <a
              class="branch-row"
              href={editUrl(deployment.branch)}
              onclick={handleBranchClick}
              data-sveltekit-preload-data="hover"
            >
              <span class="font-mono truncate">
                {deployment.branch || sourceBranch}
              </span>
            </a>
          {/each}
        </div>

        <div class="separator"></div>

        {#if showNewBranchInput}
          <div class="form">
            <label class="form-label" for="new-branch-name">
              New branch name
            </label>
            <!-- svelte-ignore a11y_autofocus -->
            <input
              id="new-branch-name"
              class="branch-input"
              type="text"
              value={branchName}
              oninput={handleBranchNameInput}
              onkeydown={handleKeydown}
              placeholder="branch-name"
              autofocus
            />
            <div class="form-actions">
              <Button
                type="ghost"
                disabled={isStarting}
                onClick={handleCancelNewBranch}
              >
                Cancel
              </Button>
              <Button
                type="primary"
                disabled={!branchName.trim() || isStarting}
                loading={isStarting}
                loadingCopy="Starting..."
                onClick={handleCreate}
              >
                Create &amp; edit
              </Button>
            </div>
          </div>
        {:else}
          <button
            class="new-branch-btn"
            onclick={() => (showNewBranchInput = true)}
          >
            <PlusIcon size="14" />
            <span>New branch&hellip;</span>
          </button>
        {/if}
      {:else}
        <div class="form">
          <label class="form-label" for="new-branch-name">Branch name</label>
          <!-- svelte-ignore a11y_autofocus -->
          <input
            id="new-branch-name"
            class="branch-input"
            type="text"
            value={branchName}
            oninput={handleBranchNameInput}
            onkeydown={handleKeydown}
            placeholder="branch-name"
            autofocus
          />
          <div class="form-actions">
            <Button
              type="primary"
              disabled={!branchName.trim() || isStarting}
              loading={isStarting}
              loadingCopy="Starting..."
              onClick={handleCreate}
            >
              Create &amp; edit
            </Button>
          </div>
        </div>
      {/if}
    </Dialog.Content>
  </Dialog.Root>
{/if}

<style lang="postcss">
  /* Header subtitle (Dialog.Description) — overrides muted default to read
     as a single inline line with the source-branch chip. */
  :global(.dlg-subtitle) {
    @apply !text-xs !text-fg-secondary whitespace-nowrap;
  }

  .branch-chip {
    @apply ml-1 inline-flex items-center gap-x-1 align-[-2px];
    @apply px-1.5 py-px rounded;
    @apply text-[11.5px] font-mono font-medium text-fg-primary;
    @apply bg-surface-subtle border border-border;
  }

  .branch-chip :global(svg) {
    @apply text-fg-muted shrink-0;
  }

  .banner {
    @apply mt-3 rounded-sm px-2.5 py-2;
    @apply text-xs text-yellow-800 bg-yellow-50 border border-yellow-200;
  }

  .branch-list {
    @apply flex flex-col mt-3 -mx-1.5;
  }

  .branch-row {
    @apply flex items-center gap-x-2 rounded-md px-3 py-2 text-sm;
    @apply text-fg-primary hover:bg-surface-hover hover:text-fg-accent;
    @apply cursor-pointer outline-none no-underline;
  }

  .separator {
    @apply my-2 h-px bg-border;
  }

  .new-branch-btn {
    @apply flex w-full items-center gap-x-2 rounded-md px-3 py-2 text-sm font-medium;
    @apply -mx-1.5 text-primary-600 hover:bg-surface-hover cursor-pointer;
  }

  .form {
    @apply flex flex-col gap-y-2 mt-3;
  }

  .form-label {
    @apply text-xs font-medium text-fg-primary;
  }

  .branch-input {
    @apply w-full text-sm font-mono px-3 py-2 rounded-md border border-gray-300;
    @apply focus:outline-none focus:ring-2 focus:ring-primary-500/30 focus:border-primary-500;
  }

  .form-actions {
    @apply flex items-center justify-end gap-x-2 mt-2;
  }
</style>
