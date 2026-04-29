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
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { Tabs as TabsPrimitive } from "bits-ui";
  import { CheckIcon, ChevronDownIcon, GitBranchIcon } from "lucide-svelte";
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
  let currentTab: "existing" | "new" = "existing";
  let selectedBranchId = "";
  let dropdownOpen = false;

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

  // Most-recently-updated branch — used for the "latest" tag and as the
  // default selection in the Existing-branch dropdown.
  $: latestDeployment = ownDeployments[0];
  $: latestBranchId = latestDeployment?.id ?? "";

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

  // Reset state every time the dialog opens. Default to the Existing tab
  // when the user has branches (most users come back to resume); preselect
  // the most-recent branch so "Continue editing" is one click away.
  $: if (open) {
    branchName = "";
    currentTab = hasOwnSessions ? "existing" : "new";
    selectedBranchId = latestBranchId;
  }

  $: selectedDeployment = ownDeployments.find(
    (d) => d.id === selectedBranchId,
  );

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

  function handleResume() {
    if (!selectedDeployment) return;
    requestSkipBranchInjection();
    open = false;
    void goto(editUrl(selectedDeployment.branch));
  }

  function handleNameKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      e.preventDefault();
      void handleCreate();
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

    <Dialog.Content class="!max-w-[480px] gap-0 p-0">
      <Dialog.Header class="space-y-1 px-6 pt-6">
        <Dialog.Title class="text-lg font-semibold tracking-tight">
          Start editing
        </Dialog.Title>
        <Dialog.Description class="text-[13px] text-fg-secondary leading-snug">
          {#if hasOwnSessions}
            Pick up an existing branch, or create a new one from
            <code class="dlg-code">{sourceBranch}</code>.
          {:else}
            We'll create a branch from
            <code class="dlg-code">{sourceBranch}</code>
            for your edits.
          {/if}
        </Dialog.Description>
      </Dialog.Header>

      {#if activeBranchIsNonEditable}
        <div class="banner">
          This branch is read-only. Pick another branch or create a new one.
        </div>
      {/if}

      {#if hasOwnSessions}
        <TabsPrimitive.Root bind:value={currentTab} class="contents">
          <TabsPrimitive.List class="seg-list">
            <TabsPrimitive.Trigger value="existing" class="seg-trigger">
              Existing branch
            </TabsPrimitive.Trigger>
            <TabsPrimitive.Trigger value="new" class="seg-trigger">
              New branch
            </TabsPrimitive.Trigger>
          </TabsPrimitive.List>

          <TabsPrimitive.Content value="existing" class="tab-body">
            <label class="form-label" for="existing-branch">Branch</label>
            <DropdownMenu.Root bind:open={dropdownOpen}>
              <DropdownMenu.Trigger>
                {#snippet child({ props })}
                  <button
                    {...props}
                    id="existing-branch"
                    class="branch-select"
                    class:open={dropdownOpen}
                    type="button"
                  >
                    <span class="select-left">
                      <GitBranchIcon
                        size="14"
                        class="text-fg-muted shrink-0"
                      />
                      <span class="select-name">
                        {selectedDeployment?.branch ?? sourceBranch}
                      </span>
                      {#if selectedDeployment && selectedDeployment.id === latestBranchId}
                        <span class="latest-tag">latest</span>
                      {/if}
                    </span>
                    <ChevronDownIcon
                      size="14"
                      class="text-fg-muted shrink-0 transition-transform {dropdownOpen
                        ? 'rotate-180'
                        : ''}"
                    />
                  </button>
                {/snippet}
              </DropdownMenu.Trigger>
              <DropdownMenu.Content
                align="start"
                sameWidth
                class="max-h-[280px] overflow-y-auto"
              >
                {#each ownDeployments as deployment (deployment.id)}
                  {@const isSelected = deployment.id === selectedBranchId}
                  {@const isLatest = deployment.id === latestBranchId}
                  <DropdownMenu.Item
                    onclick={() => (selectedBranchId = deployment.id ?? "")}
                    class="branch-option"
                  >
                    <GitBranchIcon
                      size="13"
                      class="text-fg-muted shrink-0"
                    />
                    <span class="font-mono text-[13px] truncate flex-1">
                      {deployment.branch || sourceBranch}
                    </span>
                    {#if isLatest}
                      <span class="latest-tag">latest</span>
                    {/if}
                    {#if isSelected}
                      <CheckIcon
                        size="13"
                        class="text-primary-600 shrink-0 ml-1"
                      />
                    {/if}
                  </DropdownMenu.Item>
                {/each}
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          </TabsPrimitive.Content>

          <TabsPrimitive.Content value="new" class="tab-body">
            <label class="form-label" for="new-branch-name">Branch name</label>
            <!-- svelte-ignore a11y_autofocus -->
            <input
              id="new-branch-name"
              class="branch-input"
              type="text"
              value={branchName}
              oninput={handleBranchNameInput}
              onkeydown={handleNameKeydown}
              placeholder="branch-name"
              autofocus
            />
          </TabsPrimitive.Content>
        </TabsPrimitive.Root>
      {:else}
        <div class="tab-body" style:margin-top="20px">
          <label class="form-label" for="new-branch-name">Branch name</label>
          <!-- svelte-ignore a11y_autofocus -->
          <input
            id="new-branch-name"
            class="branch-input"
            type="text"
            value={branchName}
            oninput={handleBranchNameInput}
            onkeydown={handleNameKeydown}
            placeholder="branch-name"
            autofocus
          />
        </div>
      {/if}

      <div class="dlg-footer">
        <Button type="secondary" onClick={() => (open = false)}>Cancel</Button>
        {#if hasOwnSessions && currentTab === "existing"}
          <Button
            type="primary"
            disabled={!selectedBranchId || isStarting}
            onClick={handleResume}
          >
            Continue editing
          </Button>
        {:else}
          <Button
            type="primary"
            disabled={!branchName.trim() || isStarting}
            loading={isStarting}
            loadingCopy="Starting..."
            onClick={handleCreate}
          >
            Create &amp; edit
          </Button>
        {/if}
      </div>
    </Dialog.Content>
  </Dialog.Root>
{/if}

<style lang="postcss">
  /* Inline monospace for branch names in subtitle prose */
  :global(.dlg-code) {
    @apply font-mono text-[12.5px] text-fg-primary bg-transparent px-0;
  }

  /* Banner for non-editable active branch */
  .banner {
    @apply mx-6 mt-4 rounded-md px-3 py-2;
    @apply text-xs text-yellow-800 bg-yellow-50 border border-yellow-200;
  }

  /* Segmented tab control — Di's pattern (gray pill, white active) */
  :global(.seg-list) {
    @apply mx-6 mt-5 inline-flex p-0.5 gap-0.5 self-start;
    @apply bg-gray-100 rounded-lg;
  }

  :global(.seg-trigger) {
    @apply px-3.5 py-1.5 rounded-md;
    @apply text-[13px] font-medium text-fg-secondary;
    @apply transition-colors cursor-pointer;
    @apply focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500/30;
  }

  :global(.seg-trigger[data-state="active"]) {
    @apply bg-surface text-fg-primary shadow-sm;
  }

  :global(.seg-trigger:not([data-state="active"]):hover) {
    @apply text-fg-primary;
  }

  /* Body — locks 16px gap from tabs (or subtitle when no tabs) */
  :global(.tab-body) {
    @apply px-6 pt-4 pb-1 flex flex-col gap-1.5;
  }

  .form-label {
    @apply text-[13px] font-medium text-fg-primary mb-1;
  }

  .branch-input {
    @apply w-full font-mono text-[13.5px] px-3 py-2.5;
    @apply bg-surface border border-gray-300 rounded-lg;
    @apply text-fg-primary placeholder:text-fg-muted;
    @apply focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500;
    @apply transition-shadow;
  }

  /* Existing-branch dropdown trigger */
  .branch-select {
    @apply flex items-center justify-between gap-2 w-full;
    @apply px-3 py-2.5 rounded-lg;
    @apply bg-surface border border-gray-300 text-left;
    @apply hover:bg-surface-hover transition-colors cursor-pointer;
    @apply focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500;
  }

  .branch-select.open {
    @apply ring-2 ring-primary-500/20 border-primary-500;
  }

  .select-left {
    @apply flex items-center gap-2 min-w-0 flex-1;
  }

  .select-name {
    @apply font-mono text-[13px] text-fg-primary truncate;
  }

  .latest-tag {
    @apply font-sans text-[10.5px] font-medium uppercase tracking-wider;
    @apply text-fg-muted shrink-0;
  }

  :global(.branch-option) {
    @apply flex items-center gap-2 px-2 py-1.5 cursor-pointer;
  }

  /* Footer separator and right-aligned button row */
  .dlg-footer {
    @apply flex items-center justify-end gap-2;
    @apply px-6 py-4 mt-6 border-t border-border;
  }
</style>
