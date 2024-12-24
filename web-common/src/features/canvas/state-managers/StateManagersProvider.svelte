<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { setContext } from "svelte";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let canvasName: string;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({
    queryClient,
    canvasName,
  });

  setContext(DEFAULT_STORE_KEY, stateManagers);
</script>

<slot />
