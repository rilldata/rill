<script lang="ts">
  import StatusTextCell from "@rilldata/web-common/features/resources/cells/StatusTextCell.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let resource: V1Resource;

  type Tone = "default" | "muted" | "success" | "destructive";

  $: ({ label, tone } = describeStatus(resource));

  function describeStatus(r: V1Resource): { label: string; tone: Tone } {
    const last = r.report?.state?.executionHistory?.[0];
    if (!last) return { label: "Pending", tone: "muted" };
    if (last.errorMessage) return { label: "Error", tone: "destructive" };
    return { label: "OK", tone: "success" };
  }
</script>

<StatusTextCell {label} {tone} />
