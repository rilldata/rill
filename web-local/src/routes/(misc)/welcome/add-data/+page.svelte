<script lang="ts">
  import AddData from "@rilldata/web-common/features/add-data/AddData.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

  $: ({ instanceId } = $runtime);
</script>

<div>
  <div class="text-base font-semibold text-fg-secondary">Getting started</div>
  <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  <div class="w-[900px] h-[600px] mt-4">
    <AddData config={{ instanceId }} />
  </div>
</div>
