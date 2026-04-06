<script lang="ts">
  import ResourceErrorMessage from "@rilldata/web-common/features/projects/status/ResourceErrorMessage.svelte";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";

  let { message }: { message: string } = $props();
</script>

<ResourceErrorMessage
  {message}
  status={V1ReconcileStatus.RECONCILE_STATUS_IDLE}
/>
