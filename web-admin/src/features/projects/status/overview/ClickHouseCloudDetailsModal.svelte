<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
  } from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let open = false;
  export let serviceName: string | undefined = undefined;
  export let status: string | undefined = undefined;
  export let provider: string | undefined = undefined;
  export let region: string | undefined = undefined;
  export let tier: string | undefined = undefined;
  export let minMemoryGb: number | undefined = undefined;
  export let maxMemoryGb: number | undefined = undefined;
  export let replicas: number | undefined = undefined;

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
</script>

<Dialog bind:open>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>ClickHouse Cloud Service Details</DialogTitle>
    </DialogHeader>

    <div class="detail-grid">
      <div class="detail-row">
        <span class="detail-label">Service</span>
        <span class="detail-value">{serviceName ?? "—"}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Status</span>
        <span class="detail-value {statusColor(status)} font-medium capitalize">
          {status ?? "—"}
        </span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Provider</span>
        <span class="detail-value uppercase">{provider ?? "—"}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Region</span>
        <span class="detail-value">{region ?? "—"}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Tier</span>
        <span class="detail-value capitalize">{tier ?? "—"}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Memory</span>
        <span class="detail-value">
          {#if minMemoryGb != null && maxMemoryGb != null}
            {minMemoryGb} GB – {maxMemoryGb} GB
          {:else}
            —
          {/if}
        </span>
      </div>
      <div class="detail-row">
        <span class="detail-label">Replicas</span>
        <span class="detail-value">{replicas ?? "—"}</span>
      </div>
    </div>

    <DialogFooter>
      <Button type="secondary" onClick={() => (open = false)}>Close</Button>
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
