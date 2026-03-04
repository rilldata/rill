<script lang="ts">
  import { page } from "$app/stores";
  import AddData from "@rilldata/web-common/features/add-data/AddData.svelte";

  const initSchemaName = ($page.state as any).schema;
  const initConnectorName = ($page.state as any).connector;
</script>

<div class="w-[900px]">
  <AddData {initSchemaName} {initConnectorName} />
</div>
