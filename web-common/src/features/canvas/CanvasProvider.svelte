<script lang="ts">
  import CanvasLoadingState from "@rilldata/web-common/features/canvas/CanvasLoadingState.svelte";
  import CanvasInitialization from "./CanvasInitialization.svelte";

  export let canvasName: string;
  export let instanceId: string;
  export let showBanner = false;
  export let projectId: string | undefined = undefined;
</script>

<CanvasInitialization
  {canvasName}
  {instanceId}
  {projectId}
  {showBanner}
  let:ready
  let:errorMessage
  let:isLoading
  let:isReconciling
>
  <CanvasLoadingState {ready} {errorMessage} {isLoading} {isReconciling}>
    <slot />
  </CanvasLoadingState>
</CanvasInitialization>
