<script lang="ts">
  import { page } from "$app/stores";
  import AddData from "@rilldata/web-common/features/add-data/AddData.svelte";

  const initSchemaName = ($page.state as any).schema;
  const initConnectorName = ($page.state as any).connector;
</script>

<div>
  <div class="text-base font-semibold text-fg-secondary">Getting started</div>
  <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  <div class="w-[900px] h-[600px] mt-4">
    <AddData {initSchemaName} {initConnectorName} />
  </div>
</div>
