<script lang="ts">
  import { setContext } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let metricsViewName: string;
  export let exploreName: string;
  export let visualEditing = false;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({
    queryClient,
    metricsViewName,
    exploreName,
  });
  setContext(DEFAULT_STORE_KEY, stateManagers);

  // Our state management was not built around the ability to arbitrarily change the explore or metrics view name
  // This needs to change, but this is a workaround for now
  $: if (visualEditing) {
    stateManagers.metricsViewName.set(metricsViewName);
  }
</script>

<slot />
