<script lang="ts">
  import {
    createTimeControlStore,
    DEFAULT_TIME_STORE_KEY,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { setContext } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let metricsViewName: string;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({ queryClient, metricsViewName });
  setContext(DEFAULT_STORE_KEY, stateManagers);

  const timeControlsStore = createTimeControlStore(stateManagers);
  setContext(DEFAULT_TIME_STORE_KEY, timeControlsStore);

  $: {
    // update metrics view name
    stateManagers.setMetricsViewName(metricsViewName);
  }
</script>

<slot />
