<script lang="ts">
  import { page } from "$app/stores";
  import { setContext } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let metricsViewName: string;

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({
    queryClient,
    metricsViewName,
    extraKeyPrefix:
      orgName && projectName ? `__${orgName}__${projectName}` : "",
  });
  setContext(DEFAULT_STORE_KEY, stateManagers);

  $: {
    // update metrics view name
    stateManagers.setMetricsViewName(metricsViewName);
  }
</script>

<slot />
