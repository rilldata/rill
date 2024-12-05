<script lang="ts">
  import StateManagersProvider from "@rilldata/web-common/features/canvas/state-managers/StateManagersProvider.svelte";
  import CanvasStateProvider from "@rilldata/web-common/features/canvas/stores/CanvasStateProvider.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type { V1Canvas } from "@rilldata/web-common/runtime-client";
  import Canvas from "./Canvas.svelte";
  import CanvasThemeProvider from "./CanvasThemeProvider.svelte";

  export let canvasName: string;
  export let canvasResource: V1Canvas;
  export let fileArtifact: FileArtifact;
</script>

{#key canvasName}
  <StateManagersProvider {canvasName} {canvasResource}>
    <CanvasStateProvider>
      <CanvasThemeProvider>
        <Canvas {fileArtifact} />
      </CanvasThemeProvider>
    </CanvasStateProvider>
  </StateManagersProvider>
{/key}
