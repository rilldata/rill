<script lang="ts">
  import { getBusinessModel } from "./business-model";

  const businessModel = getBusinessModel();
  const { dashboardStore, metricsViewName } = businessModel;
</script>

<div>The dashboard is: {$metricsViewName}</div>
<div>
  The filters are: <pre>{JSON.stringify($dashboardStore.filters, null, 2)}</pre>
</div>
