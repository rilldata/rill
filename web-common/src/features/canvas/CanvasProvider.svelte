<script lang="ts">
  import CanvasLoadingState from "@rilldata/web-common/features/canvas/CanvasLoadingState.svelte";
  import CanvasInitialization from "./CanvasInitialization.svelte";

  export let canvasName: string;
  export let instanceId: string;
  export let showBanner = false;
  export let projectId: string | undefined = undefined;
  // When set, relative time ranges are anchored at this time instead of now/latest
  // (e.g. for scheduled report exports).
  export let executionTime: string | undefined = undefined;
  // See CanvasInitialization for the isolated contract.
  export let isolated = false;
</script>

<CanvasInitialization
  {canvasName}
  {instanceId}
  {projectId}
  {showBanner}
  {executionTime}
  {isolated}
  let:ready
  let:reconcileErrorMessage
  let:isLoading
  let:isReconciling
  let:queryError
>
  <CanvasLoadingState
    {ready}
    errorMessage={reconcileErrorMessage}
    {isLoading}
    {isReconciling}
    {queryError}
  >
    <slot />
  </CanvasLoadingState>
</CanvasInitialization>
