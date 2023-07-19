<script lang="ts">
  import { setContext } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let metricsViewName: string;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({ queryClient, metricsViewName });
  setContext(DEFAULT_STORE_KEY, stateManagers);

  $: {
    // update metrics view name
    stateManagers.setMetricsViewName(metricsViewName);
  }
</script>

<slot />
