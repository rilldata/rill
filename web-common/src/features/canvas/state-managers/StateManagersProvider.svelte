<script lang="ts">
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import type { V1Canvas } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { setContext } from "svelte";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let canvasName: string;
  export let canvasResource: V1Canvas;
  export let fileArtifact: FileArtifact;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({
    queryClient,
    canvasName,
    canvasResource,
    fileArtifact,
  });

  setContext(DEFAULT_STORE_KEY, stateManagers);
</script>

<slot />
