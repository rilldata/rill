<script lang="ts">
  import { page } from "$app/stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createStateManagers, stateManagersContext } from "./state-managers";

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

  stateManagersContext.set(stateManagers);

  $: {
    // update metrics view name
    stateManagers.setMetricsViewName(metricsViewName);
  }
</script>

<slot />
