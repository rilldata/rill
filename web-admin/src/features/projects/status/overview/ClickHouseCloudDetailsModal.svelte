<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
  } from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { CANONICAL_ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import { createEventDispatcher } from "svelte";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let serviceName: string | undefined = undefined;
  export let status: string | undefined = undefined;
  export let provider: string | undefined = undefined;
  export let region: string | undefined = undefined;
  export let minMemoryGb: number | undefined = undefined;
  export let maxMemoryGb: number | undefined = undefined;
  export let replicas: number | undefined = undefined;

  const dispatch = createEventDispatcher();

  let syncing = false;
  let syncError: string | null = null;

  // Local overrides from sync response (takes precedence over props)
  let syncedData: Record<string, unknown> | null = null;
  $: displayServiceName =
    (syncedData?.cloud_service_name as string) ?? serviceName;
  $: displayStatus = (syncedData?.cloud_status as string) ?? status;
  $: displayProvider = (syncedData?.cloud_provider as string) ?? provider;
  $: displayRegion = (syncedData?.cloud_region as string) ?? region;
  $: displayMinMemory =
    (syncedData?.cloud_min_memory_gb as number) ?? minMemoryGb;
  $: displayMaxMemory =
    (syncedData?.cloud_max_memory_gb as number) ?? maxMemoryGb;
  $: displayReplicas = (syncedData?.cloud_num_replicas as number) ?? replicas;

  function statusColor(s: string | undefined): string {
    switch (s) {
      case "running":
        return "text-green-600";
      case "idle":
        return "text-amber-600";
      case "stopped":
        return "text-red-600";
      default:
        return "text-fg-secondary";
    }
  }

  async function syncNow() {
    syncing = true;
    syncError = null;
    try {
      const resp = await fetch(
        `${CANONICAL_ADMIN_URL}/v1/clickhouse-cloud/sync`,
        {
          method: "POST",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ org: organization, project }),
        },
      );
      if (!resp.ok) {
        syncError = await resp.text();
      } else {
        const data = await resp.json();
        syncedData = data;
        dispatch("synced", { maxMemoryGb: data.cloud_max_memory_gb });
      }
    } catch (e) {
      syncError = `${e}`;
    } finally {
      syncing = false;
    }
  }
</script>

<Dialog bind:open>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>ClickHouse Cloud Service Details</DialogTitle>
    </DialogHeader>

    <div class="detail-grid">
      <div class="detail-row">
        <span class="detail-label">Service</span>
        <span class="detail-value">{displayServiceName ?? "—"}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Status</span>
        <span
          class="detail-value {statusColor(
            displayStatus,
          )} font-medium capitalize"
        >
          {displayStatus ?? "—"}
        </span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Provider</span>
        <span class="detail-value uppercase">{displayProvider ?? "—"}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Region</span>
        <span class="detail-value">{displayRegion ?? "—"}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Memory</span>
        <span class="detail-value">
          {#if displayMinMemory != null && displayMaxMemory != null}
            {displayMinMemory} GB – {displayMaxMemory} GB
          {:else}
            —
          {/if}
        </span>
      </div>
      <div class="detail-row">
        <span class="detail-label">vCPU</span>
        <span class="detail-value">
          {#if displayMinMemory != null && displayMaxMemory != null}
            {displayMinMemory / 4} – {displayMaxMemory / 4}
          {:else}
            —
          {/if}
        </span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Replicas</span>
        <span class="detail-value">{displayReplicas ?? "—"}</span>
      </div>
    </div>

    {#if syncError}
      <p class="text-red-600 text-xs mt-2">{syncError}</p>
    {/if}

    <DialogFooter>
      <Button type="secondary" onClick={() => (open = false)}>Close</Button>
      <Button type="primary" onClick={syncNow} disabled={syncing}>
        {syncing ? "Syncing..." : "Sync Now"}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>

<style lang="postcss">
  .detail-grid {
    @apply flex flex-col;
  }
  .detail-row {
    @apply flex items-center py-2 border-b border-border;
  }
  .detail-row:last-child {
    @apply border-b-0;
  }
  .detail-label {
    @apply text-sm text-fg-secondary w-28 shrink-0;
  }
  .detail-value {
    @apply text-sm text-fg-primary;
  }
</style>
