# Using the Business Model Provider
The BusinessModelProvider provides an object with all dependent services for the Rill UI's business model logic.
Here is an example of using it:
```svelte
<script lang="ts">
  import { getBusinessModel } from "./business-model";

  const businessModel = getBusinessModel();
  const { dashboardStore, metricsViewName } = businessModel;
</script>

<div>The dashboard is: {$metricsViewName}</div>
<div>
  The filters are: <pre>{JSON.stringify($dashboardStore.filters, null, 2)}</pre>
</div>

```