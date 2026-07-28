<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetProjectFileSyncStatus,
    createAdminServiceSyncProjectFilesToGit,
    createAdminServiceUpdateProjectFileSyncSchedule,
    getAdminServiceGetProjectFileSyncStatusQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import OverviewCard from "@rilldata/web-common/features/projects/status/overview/OverviewCard.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import type { AxiosError } from "axios";

  export let organization: string;
  export let project: string;

  const autoSyncOptions = [
    { value: "0", label: "Disabled" },
    { value: "3600", label: "Every hour" },
    { value: "21600", label: "Every 6 hours" },
    { value: "86400", label: "Every day" },
  ];

  let dialogOpen = false;
  let syncing = false;
  // lastSyncedOn at the time the sync was requested; the job bumps it when it completes successfully.
  let syncBaselineSyncedOn: string | undefined;

  // Only show this card for Git-connected projects (managed-git projects sync transparently elsewhere).
  $: projectResp = createAdminServiceGetProject(organization, project);
  $: isGitConnected =
    !!$projectResp.data?.project?.gitRemote &&
    !$projectResp.data?.project?.managedGitId;
  $: primaryBranch =
    $projectResp.data?.project?.primaryBranch ?? "the main branch";

  // Poll while a sync is in flight so the card reflects the outcome without a manual refresh.
  $: statusQuery = createAdminServiceGetProjectFileSyncStatus(
    organization,
    project,
    {
      query: {
        enabled: isGitConnected,
        refetchInterval: syncing ? 2000 : false,
      },
    },
  );

  $: status = $statusQuery.data;
  $: unsyncedCount = status?.unsyncedCount ?? 0;
  $: hasChanges = unsyncedCount > 0;
  $: lastSyncedOn = status?.lastSyncedOn ? new Date(status.lastSyncedOn) : null;
  $: lastSyncError = status?.lastSyncError ?? "";
  $: lastSyncWarning = status?.lastSyncWarning ?? "";
  $: autoSyncInterval = autoSyncOptions.some(
    (o) => o.value === status?.syncIntervalSeconds,
  )
    ? (status?.syncIntervalSeconds ?? "0")
    : "0";

  const syncMutation = createAdminServiceSyncProjectFilesToGit();
  const scheduleMutation = createAdminServiceUpdateProjectFileSyncSchedule();

  // Stop polling once a sync settles: everything synced, the job completed successfully but files were
  // edited while it ran (lastSyncedOn advanced with a nonzero count), or the job reported a failure
  // (requesting a sync clears the previous error, so a non-empty error here is this sync's outcome).
  $: if (
    syncing &&
    $statusQuery.isFetched &&
    (unsyncedCount === 0 ||
      lastSyncError ||
      (status?.lastSyncedOn && status.lastSyncedOn !== syncBaselineSyncedOn))
  ) {
    syncing = false;
    if (lastSyncError) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to sync user files: ${lastSyncError}`,
      });
    } else if (lastSyncWarning) {
      // Requesting a sync clears the server-side warning, so a non-empty warning here is this sync's outcome.
      eventBus.emit("notification", {
        type: "warning",
        message: `User files synced, but ${lastSyncWarning}`,
      });
    }
  }

  function invalidateStatus() {
    return queryClient.invalidateQueries({
      queryKey: getAdminServiceGetProjectFileSyncStatusQueryKey(
        organization,
        project,
      ),
    });
  }

  async function sync() {
    try {
      await $syncMutation.mutateAsync({
        org: organization,
        project,
        data: {},
      });
      dialogOpen = false;
      syncBaselineSyncedOn = status?.lastSyncedOn;
      // Refetch before enabling the settle check: requesting a sync clears the server-side error,
      // and the check must not mistake a cached error from a previous sync for this sync's outcome.
      await invalidateStatus();
      syncing = true;
      eventBus.emit("notification", {
        type: "success",
        message: "Syncing user files to Git",
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        type: "error",
        message:
          axiosError.response?.data?.message ?? "Failed to sync user files",
      });
    }
  }

  async function updateAutoSync(value: string) {
    if (value === autoSyncInterval) return;
    try {
      await $scheduleMutation.mutateAsync({
        org: organization,
        project,
        data: { syncIntervalSeconds: value },
      });
      await invalidateStatus();
      eventBus.emit("notification", {
        type: "success",
        message:
          value === "0" ? "Auto-sync disabled" : "Auto-sync schedule updated",
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        type: "error",
        message:
          axiosError.response?.data?.message ??
          "Failed to update auto-sync schedule",
      });
    }
  }
</script>

{#if isGitConnected}
  <ProjectAccessControls {organization} {project}>
    <svelte:fragment slot="manage-project">
      <OverviewCard title="User files">
        <div class="flex flex-col gap-2 text-sm text-gray-600">
          {#if hasChanges}
            <p>
              {unsyncedCount} unsynced change{unsyncedCount === 1 ? "" : "s"}.
              Sync to commit alerts, reports, and personal dashboards into your
              project's Git repo.
            </p>
          {:else}
            <p>All user files are synced to Git.</p>
          {/if}
          {#if lastSyncError}
            <p class="text-red-600">Last sync failed: {lastSyncError}</p>
          {:else if lastSyncWarning}
            <p class="text-amber-600">Last sync: {lastSyncWarning}</p>
          {/if}
          {#if lastSyncedOn}
            <p class="text-gray-500">Last synced {timeAgo(lastSyncedOn)}</p>
          {/if}
        </div>

        <div slot="header-right" class="flex items-center gap-3">
          <Select
            id="user-files-auto-sync"
            ariaLabel="Auto-sync interval"
            size="sm"
            options={autoSyncOptions.map((o, i) => ({
              ...o,
              label:
                i === 0
                  ? "Auto-sync: off"
                  : `Auto-sync: ${o.label.toLowerCase()}`,
            }))}
            value={autoSyncInterval}
            onChange={updateAutoSync}
          />
          <AlertDialog bind:open={dialogOpen}>
            <AlertDialogTrigger>
              {#snippet child({ props })}
                <Button
                  {...props}
                  type="secondary"
                  loading={syncing}
                  loadingCopy="Syncing…"
                  disabled={!hasChanges}
                >
                  Sync to Git
                </Button>
              {/snippet}
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Sync user files to Git?</AlertDialogTitle>
                <AlertDialogDescription>
                  This commits {unsyncedCount} file{unsyncedCount === 1
                    ? ""
                    : "s"} directly to <code>{primaryBranch}</code>. It syncs
                  user files from all members of this project, not just yours.
                  If any of these files were also edited directly in Git, the
                  sync overwrites those edits (previous versions remain in Git
                  history). If the branch is protected and rejects the push, the
                  sync will fail and your files stay staged in Rill.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <Button type="tertiary" onClick={() => (dialogOpen = false)}>
                  Cancel
                </Button>
                <Button
                  type="primary"
                  onClick={sync}
                  loading={$syncMutation.isPending}
                >
                  Sync
                </Button>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </OverviewCard>
    </svelte:fragment>
  </ProjectAccessControls>
{/if}
