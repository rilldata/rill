<script lang="ts">
  import { MenuItem } from "@rilldata/web-common/components/menu";

  export let chartName: string;

  function handleDeleteChart() {
    console.log("delete chart", chartName);
  }
</script>

<MenuItem icon on:select={handleDeleteChart} propogateSelect={false}>
  Delete chart
</MenuItem>
