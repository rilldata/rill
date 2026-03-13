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
  $: label = hasOwnSessions ? "Open editor" : "Edit";
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
    if (directEditHref) {
      e.preventDefault();
      e.stopPropagation();
      open = false;
      requestSkipBranchInjection();
      void goto(directEditHref);
    }
    // Otherwise the Popover.Trigger handles the click naturally
  }

  function handleNavigate(branch: string | undefined) {
    open = false;
    requestSkipBranchInjection();
    void goto(editUrl(branch));
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

<Popover.Root bind:open>
  <Popover.Trigger asChild let:builder>
    <Button
      type="secondary"
      href={directEditHref}
      builders={[builder]}
      disabled={isStarting || isLoading}
      loading={isStarting}
      loadingCopy="Starting..."
      onClick={handleButtonClick}
    >
      {label}
    </Button>
  </Popover.Trigger>

  <Popover.Content class="w-[280px]" align="end" sideOffset={8} padding="2">
    <div class="flex flex-col">
      {#if hasOwnSessions}
        <!-- List of active branches -->
        {#each ownDeployments as deployment (deployment.id)}
          <a
            class="branch-item"
            href={editUrl(deployment.branch)}
            on:click|preventDefault={() => handleNavigate(deployment.branch)}
          >
            <span
              class="inline-block size-1.5 rounded-full flex-none {statusDot(
                deployment.status,
              )}"
            />
            <span class="font-mono text-xs truncate">
              {deployment.branch || "main"}
            </span>
          </a>
        {/each}

        <!-- Separator + New branch -->
        <div class="border-t border-gray-200 my-1"></div>

        {#if showNewBranchInput}
          <div class="flex flex-col gap-y-1.5 px-1 pb-1">
            <input
              class="branch-input"
              type="text"
              bind:value={branchName}
              on:keydown={handleKeydown}
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
            class="branch-item text-primary-600"
            on:click={() => {
              showNewBranchInput = true;
            }}
          >
            <PlusIcon size="14" />
            <span class="text-sm">New branch...</span>
          </button>
        {/if}
      {:else}
        <!-- No sessions: show branch name input directly -->
        <div class="flex flex-col gap-y-1.5 px-1 pb-1">
          <label class="text-xs text-fg-secondary">Branch name</label>
          <input
            class="branch-input"
            type="text"
            bind:value={branchName}
            on:keydown={handleKeydown}
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
    </div>
  </Popover.Content>
</Popover.Root>

<style lang="postcss">
  .branch-item {
    @apply flex items-center gap-x-2 px-2 py-1.5 rounded text-left w-full;
    @apply hover:bg-surface-hover transition-colors cursor-pointer;
  }

  .branch-input {
    @apply w-full text-sm font-mono px-2 py-1 rounded border border-gray-300;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
  }
</style>
