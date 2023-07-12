# Using the State Managers Provider
The StateManagersProvider provides an object with all dependent services for the Rill UI's business model logic.
Here is an example of using it:
```svelte
<script lang="ts">
  import { getStateManagers } from "./state-managers";

  const stateManagers = getStateManagers();
  const { dashboardStore, metricsViewName } = stateManagers;
</script>

<div>The dashboard is: {$metricsViewName}</div>
<div>
  The filters are: <pre>{JSON.stringify($dashboardStore.filters, null, 2)}</pre>
</div>

```