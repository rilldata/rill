<script lang="ts">
  import StatusTextCell from "@rilldata/web-common/features/resources/cells/StatusTextCell.svelte";
  import {
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";

  export let resource: V1Resource;

  $: ({ label, tone } = describeStatus(resource));

  type Tone = "default" | "muted" | "success" | "destructive";

  function describeStatus(r: V1Resource): { label: string; tone: Tone } {
    if (r.meta?.reconcileError) return { label: "Error", tone: "destructive" };

    switch (r.meta?.reconcileStatus) {
      case V1ReconcileStatus.RECONCILE_STATUS_RUNNING:
        return { label: "Reconciling", tone: "muted" };
      case V1ReconcileStatus.RECONCILE_STATUS_PENDING:
        return { label: "Building", tone: "muted" };
      case V1ReconcileStatus.RECONCILE_STATUS_IDLE:
        return { label: "Ready", tone: "success" };
      default:
        return { label: "—", tone: "muted" };
    }
  }
</script>

<StatusTextCell {label} {tone} />
