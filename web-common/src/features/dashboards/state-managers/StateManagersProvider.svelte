<script lang="ts">
  import { page } from "$app/stores";
  import { setContext } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";

  export let metricsViewName: string;
  export let exploreName: string;
  export let visualEditing = false;

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;

  const queryClient = useQueryClient();
  const stateManagers = createStateManagers({
    queryClient,
    metricsViewName,
    exploreName,
    extraKeyPrefix:
      orgName && projectName ? `__${orgName}__${projectName}` : "",
  });
  setContext(DEFAULT_STORE_KEY, stateManagers);

  // Our state management was not built around the ability to arbitrarily change the explore or metrics view name
  // This needs to change, but this is a workaround for now
  $: if (visualEditing) {
    stateManagers.metricsViewName.set(metricsViewName);
  }
</script>

<slot />
