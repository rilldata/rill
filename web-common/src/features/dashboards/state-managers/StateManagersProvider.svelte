<script lang="ts">
  import { getContext, setContext } from "svelte";
  import { createStateManagers, DEFAULT_STORE_KEY } from "./state-managers";
  import { useExploreState } from "../stores/dashboard-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let metricsViewName: string | undefined;
  export let exploreName: string;
  export let organization: string = getContext("organization");
  export let project: string = getContext("project");
  export let visualEditing = false;

  const stateManagers = metricsViewName
    ? createStateManagers({
        queryClient,
        metricsViewName,
        exploreName,
        organization,
        project,
      })
    : undefined;

  if (stateManagers) setContext(DEFAULT_STORE_KEY, stateManagers);

  // Our state management was not built around the ability to arbitrarily change the explore or metrics view name
  // This needs to change, but this is a workaround for now
  $: if (visualEditing && stateManagers && metricsViewName) {
    stateManagers?.metricsViewName.set(metricsViewName);
  }

  $: exploreStore = useExploreState(exploreName);

  $: ready = Boolean($exploreStore) && Boolean(stateManagers);
</script>

<slot {ready} />
