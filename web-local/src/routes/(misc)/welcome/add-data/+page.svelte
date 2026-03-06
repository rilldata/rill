<script lang="ts">
  import AddData from "@rilldata/web-common/features/add-data/AddData.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
</script>

<div>
  <div class="text-base font-semibold text-fg-secondary">Getting started</div>
  <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  <div class="w-[900px] h-[600px] mt-4">
    <AddData config={{ runtimeClient }} />
  </div>
</div>
