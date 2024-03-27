<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  export let chartType = "default";
</script>

<div class="flex gap-x-2 ml-auto">
  <Button type="text" compact on:click={() => (chartType = "default")}
    >Default</Button
  >
  <Button type="text" compact on:click={() => (chartType = "bar")}>Bar</Button>
  <Button type="text" compact on:click={() => (chartType = "stacked bar")}
    >Stacked Bar</Button
  >
  <Button type="text" compact on:click={() => (chartType = "stacked area")}
    >Stacked Area</Button
  >
</div>
