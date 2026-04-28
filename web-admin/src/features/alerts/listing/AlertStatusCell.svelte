<script lang="ts">
  import StatusTextCell from "@rilldata/web-common/features/resources/cells/StatusTextCell.svelte";
  import {
    V1AssertionStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";

  export let resource: V1Resource;

  type Tone = "default" | "muted" | "success" | "destructive";

  $: ({ label, tone } = describeStatus(resource));

  function describeStatus(r: V1Resource): { label: string; tone: Tone } {
    const last = r.alert?.state?.executionHistory?.[0];
    if (!last) return { label: "Pending", tone: "muted" };

    switch (last.result?.status) {
      case V1AssertionStatus.ASSERTION_STATUS_FAIL:
        return { label: "Triggered", tone: "destructive" };
      case V1AssertionStatus.ASSERTION_STATUS_PASS:
        return { label: "OK", tone: "success" };
      case V1AssertionStatus.ASSERTION_STATUS_ERROR:
        return { label: "Error", tone: "destructive" };
      default:
        return { label: "Pending", tone: "muted" };
    }
  }
</script>

<StatusTextCell {label} {tone} />
