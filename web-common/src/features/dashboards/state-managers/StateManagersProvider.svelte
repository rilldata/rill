<script lang="ts">
  import { onMount, setContext } from "svelte";
  import {
    createStateManagers,
    DEFAULT_STORE_KEY,
    type StateManagers,
  } from "./state-managers";
  import { useExploreState } from "../stores/dashboard-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let metricsViewName: string | undefined;
  export let exploreName: string;
  export let visualEditing = false;

  let stateManagers: StateManagers | undefined = undefined;

  onMount(() => {
    if (metricsViewName) {
      stateManagers = createStateManagers({
        queryClient,
        metricsViewName,
        exploreName,
      });
      setContext(DEFAULT_STORE_KEY, stateManagers);
    }
  });

  // // Our state management was not built around the ability to arbitrarily change the explore or metrics view name
  // // This needs to change, but this is a workaround for now
  $: if (visualEditing && stateManagers && metricsViewName) {
    stateManagers?.metricsViewName.set(metricsViewName);
  }

  $: exploreStore = useExploreState(exploreName);

  $: ready = Boolean($exploreStore);
</script>

<slot {ready} />
