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
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import {
    CheckIcon,
    ChevronDownIcon,
    GitBranchIcon,
    GitBranchPlusIcon,
  } from "lucide-svelte";
  import { useDevDeployments, invalidateDeployments } from "./use-edit-session";

  export let open = false;
  export let organization: string;
  export let project: string;
  /** The branch currently being viewed (from the URL), if any. */
  export let activeBranch: string | undefined = undefined;
  /** The project's primary branch, used as the source for new branches. */
  export let primaryBranch: string | undefined = undefined;

  const user = createAdminServiceGetCurrentUser();
  const devDeployments = useDevDeployments(organization, project);
  const createMutation = createAdminServiceCreateDeployment();

  let branchName = "";
  let currentTab: "existing" | "new" = "existing";
  let selectedBranchId = "";
  let dropdownOpen = false;
  let createError = "";

  $: currentUserId = $user.data?.user?.id;
  $: deployments = $devDeployments.data?.deployments ?? [];
  $: sourceBranch = primaryBranch || "main";

  // Editable deployments owned by the current user, sorted by most recently
  // updated. Stopped/errored branches show so the user can resume or retry.
  $: ownDeployments = deployments
    .filter(
      (d) =>
        d.ownerUserId === currentUserId &&
        d.editable &&
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING &&
        d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
    )
    .sort((a, b) => (b.updatedOn ?? "").localeCompare(a.updatedOn ?? ""));

  $: latestBranchId = ownDeployments[0]?.id ?? "";
  $: hasOwnSessions = ownDeployments.length > 0;
  $: isStarting = $createMutation.isPending;

  // Banner condition: active branch is owned but not editable
  $: activeBranchIsNonEditable =
    !!activeBranch &&
    !!currentUserId &&
    deployments.some(
      (d) =>
        d.branch === activeBranch &&
        d.ownerUserId === currentUserId &&
        !d.editable,
    );

  // Reset all state every time the dialog opens.
  $: if (open) {
    resetState();
  }

  $: selectedDeployment = ownDeployments.find((d) => d.id === selectedBranchId);

  function resetState() {
    branchName = "";
    selectedBranchId = latestBranchId;
    currentTab = hasOwnSessions ? "existing" : "new";
    createError = "";
  }

  function handleTabChange(value: string) {
    currentTab = value as "existing" | "new";
    // Clear field-level errors and inputs when switching tabs so stale state
    // from one tab doesn't bleed into the other.
    branchName = "";
    createError = "";
  }

  function editUrl(branch: string | undefined): string {
    return `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  }

  // Replaces whitespace with "-" as the user types so branch names are always
  // valid. Space → "-" is a 1:1 swap, so cursor stays put.
  function handleNameInput(
    newValue: string,
    e: Event & { currentTarget: EventTarget & HTMLElement },
  ) {
    const sanitized = newValue.replace(/\s+/g, "-");
    if (sanitized !== newValue) {
      const target = e.currentTarget as HTMLInputElement;
      const cursorPos = target.selectionStart ?? sanitized.length;
      target.value = sanitized;
      target.setSelectionRange(cursorPos, cursorPos);
    }
    branchName = sanitized;
    createError = "";
  }

  async function handleCreate() {
    if (!branchName.trim()) return;
    createError = "";
    const requestedBranch = branchName.trim();

    // Front-run the obvious failure: same name as an existing branch.
    if (ownDeployments.some((d) => d.branch === requestedBranch)) {
      createError = `A branch named "${requestedBranch}" already exists.`;
      return;
    }

    try {
      const resp = await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
          branch: requestedBranch,
        },
      });
      void invalidateDeployments(organization, project);
      open = false;
      requestSkipBranchInjection();
      await goto(editUrl(resp.deployment?.branch));
    } catch (err) {
      createError =
        getRpcErrorMessage(err as any) ?? "Failed to start edit session.";
    }
  }

  function handleResume() {
    if (!selectedDeployment) return;
    requestSkipBranchInjection();
    open = false;
    void goto(editUrl(selectedDeployment.branch));
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Start editing</Dialog.Title>
      <Dialog.Description>
        {#if hasOwnSessions}
          Edit an existing branch or create a new one from
          <code class="font-mono text-fg-primary">{sourceBranch}</code>.
        {:else}
          We'll create a branch from
          <code class="font-mono text-fg-primary">{sourceBranch}</code>
          for your edits.
        {/if}
      </Dialog.Description>
    </Dialog.Header>

    {#if activeBranchIsNonEditable}
      <div
        class="rounded-md border border-yellow-200 bg-yellow-50 px-3 py-2 text-xs text-yellow-800"
      >
        This branch is read-only. Pick another branch or create a new one.
      </div>
    {/if}

    {#if hasOwnSessions}
      <Tabs value={currentTab} onValueChange={handleTabChange} class="w-full">
        <TabsList
          class="flex h-9 w-full rounded-lg border border-gray-200 bg-surface-muted p-1"
        >
          <TabsTrigger
            value="existing"
            class="flex h-7 flex-1 items-center justify-center gap-1.5 rounded-md text-sm transition-all data-[state=active]:bg-surface-overlay data-[state=active]:font-semibold data-[state=active]:shadow-sm"
          >
            <GitBranchIcon size="14" />
            Existing branch
          </TabsTrigger>
          <TabsTrigger
            value="new"
            class="flex h-7 flex-1 items-center justify-center gap-1.5 rounded-md text-sm transition-all data-[state=active]:bg-surface-overlay data-[state=active]:font-semibold data-[state=active]:shadow-sm"
          >
            <GitBranchPlusIcon size="14" />
            New branch
          </TabsTrigger>
        </TabsList>

        <TabsContent value="existing" class="mt-4 space-y-1.5">
          <span class="text-sm font-medium text-fg-primary">Branch</span>
          <DropdownMenu.Root bind:open={dropdownOpen}>
            <DropdownMenu.Trigger>
              {#snippet child({ props })}
                <button
                  {...props}
                  type="button"
                  class="flex w-full items-center justify-between gap-2 rounded-lg border border-gray-300 bg-surface px-3 py-2.5 text-left transition-colors hover:bg-surface-hover focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20"
                >
                  <span class="flex min-w-0 flex-1 items-center gap-2">
                    <GitBranchIcon size="14" class="shrink-0 text-fg-muted" />
                    <span class="truncate font-mono text-sm text-fg-primary">
                      {selectedDeployment?.branch ?? sourceBranch}
                    </span>
                    {#if selectedDeployment?.id === latestBranchId}
                      <span
                        class="shrink-0 text-[10.5px] font-medium uppercase tracking-wider text-fg-muted"
                      >
                        latest
                      </span>
                    {/if}
                  </span>
                  <ChevronDownIcon
                    size="14"
                    class="shrink-0 text-fg-muted transition-transform {dropdownOpen
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
                  class="flex cursor-pointer items-center gap-2 px-2 py-1.5"
                >
                  <GitBranchIcon size="13" class="shrink-0 text-fg-muted" />
                  <span class="flex-1 truncate font-mono text-[13px]">
                    {deployment.branch || sourceBranch}
                  </span>
                  {#if isLatest}
                    <span
                      class="shrink-0 text-[10.5px] font-medium uppercase tracking-wider text-fg-muted"
                    >
                      latest
                    </span>
                  {/if}
                  {#if isSelected}
                    <CheckIcon
                      size="13"
                      class="ml-1 shrink-0 text-primary-600"
                    />
                  {/if}
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </TabsContent>

        <TabsContent value="new" class="mt-4">
          <Input
            id="new-branch-name"
            label="Branch name"
            placeholder="branch-name"
            bind:value={branchName}
            onInput={handleNameInput}
            errors={createError || undefined}
            alwaysShowError
            capitalizeLabel={false}
            textClass="text-sm"
            fontFamily="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
            claimFocusOnMount
          />
        </TabsContent>
      </Tabs>
    {:else}
      <Input
        id="new-branch-name"
        label="Branch name"
        placeholder="branch-name"
        bind:value={branchName}
        onInput={handleNameInput}
        errors={createError || undefined}
        alwaysShowError
        capitalizeLabel={false}
        textClass="text-sm"
        fontFamily="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
        claimFocusOnMount
      />
    {/if}

    <Dialog.Footer>
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
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
