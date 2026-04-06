<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";

  export let organization: string;
  export let project: string;

  // Values are strings because the generated API types use string for int64 fields
  const CADENCE_OPTIONS = [
    { value: "3600", label: "1 hour" },
    { value: "10800", label: "3 hours" },
    { value: "21600", label: "6 hours" },
    { value: "43200", label: "12 hours" },
    { value: "86400", label: "24 hours" },
  ];

  const DEFAULT_CADENCE = "10800"; // 3 hours

  const updateProjectMutation = createAdminServiceUpdateProject();

  $: projectResp = createAdminServiceGetProject(organization, project);

  let enabled = false;
  let selectedCadence = DEFAULT_CADENCE;
  // Tracks the last-saved state so we can detect unsaved changes
  let savedEnabled = false;
  let savedCadence = DEFAULT_CADENCE;
  let formInitialized = false;

  // Sync from server on first load
  $: if ($projectResp.data?.project && !formInitialized) {
    const ttl = $projectResp.data.project.prodTtlSeconds;
    const isEnabled = !!ttl && ttl !== "0";
    enabled = isEnabled;
    savedEnabled = isEnabled;
    if (isEnabled && ttl) {
      const match = CADENCE_OPTIONS.find((o) => o.value === ttl);
      const cadence = match ? ttl : nearestCadence(ttl);
      selectedCadence = cadence;
      savedCadence = cadence;
    }
    formInitialized = true;
  }

  $: changed =
    enabled !== savedEnabled || (enabled && selectedCadence !== savedCadence);

  function nearestCadence(ttl: string): string {
    const seconds = parseInt(ttl, 10);
    if (isNaN(seconds)) return DEFAULT_CADENCE;
    let closest = CADENCE_OPTIONS[0];
    for (const opt of CADENCE_OPTIONS) {
      if (
        Math.abs(parseInt(opt.value) - seconds) <
        Math.abs(parseInt(closest.value) - seconds)
      ) {
        closest = opt;
      }
    }
    return closest.value;
  }

  async function save() {
    // Backend treats 0 as "clear TTL" (sets to NULL = disabled)
    const ttl = enabled ? selectedCadence : "0";
    try {
      await $updateProjectMutation.mutateAsync({
        org: organization,
        project: project,
        data: {
          prodTtlSeconds: ttl,
        },
      });

      void queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });

      // Update saved state immediately so the UI reflects the change
      savedEnabled = enabled;
      savedCadence = selectedCadence;

      eventBus.emit("notification", {
        message: enabled
          ? `Auto-hibernation set to ${CADENCE_OPTIONS.find((o) => o.value === selectedCadence)?.label}`
          : "Auto-hibernation disabled",
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message:
          axiosError.response?.data?.message ??
          "Failed to update auto-hibernation settings",
        type: "error",
      });
    }
  }
</script>

<div class="settings-container">
  <div class="settings-header">
    <div class="settings-title">
      <span>Auto-Hibernation</span>
      <Switch bind:checked={enabled} />
    </div>
    <div class="settings-body">
      <p>
        Automatically hibernate the project after a period of inactivity to save
        resources. Hibernated projects wake up automatically when accessed.
      </p>
      {#if enabled}
        <div class="flex items-center gap-x-3 mt-4">
          <span class="text-sm">Hibernate after</span>
          <Select
            id="auto-hibernate-cadence"
            bind:value={selectedCadence}
            options={CADENCE_OPTIONS}
            size="sm"
          />
          <span class="text-sm">of inactivity</span>
        </div>
      {/if}
    </div>
  </div>
  <div class="settings-footer">
    <div class="grow"></div>
    <Button
      disabled={!changed}
      onClick={save}
      type="primary"
      loading={$updateProjectMutation.isPending}
    >
      Save
    </Button>
  </div>
</div>

<style lang="postcss">
  .settings-container {
    @apply w-full border text-fg-secondary rounded-sm bg-surface-background;
  }

  .settings-header {
    @apply p-5;
  }

  .settings-title {
    @apply flex flex-row justify-between items-center mb-2;
    @apply text-lg font-semibold text-fg-primary;
  }

  .settings-body {
    @apply text-sm text-fg-tertiary;
  }

  .settings-footer {
    @apply flex flex-row items-center px-5 py-2;
    @apply bg-surface-subtle text-fg-tertiary text-sm border-t;
  }
</style>
