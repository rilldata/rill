<script lang="ts">
  import { getBusinessModel } from "./business-model";

  const businessModel = getBusinessModel();

  $: {
    console.log({ businessModel });
  }

  const dashStore = businessModel.dashboardStore;
  $: filters = $dashStore.filters;

  $: {
    console.log({ filters });
  }
</script>

<div>hello world</div>
