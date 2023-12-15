<script lang="ts">
  import { cubicOut } from "svelte/easing";
  import { scale } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useIsSourceUnsaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  const sourceStore = useSourceStore(sourceName);

  $: isSourceUnsavedQuery = useIsSourceUnsaved(
    $runtime.instanceId,
    sourceName,
    $sourceStore.clientYAML
  );
  $: isSourceUnsaved = $isSourceUnsavedQuery.data;
</script>

{#if isSourceUnsaved}
  <div
    transition:scale|global={{ duration: 200, easing: cubicOut }}
    class="w-1.5 h-1.5 bg-gray-300 rounded"
  />
{/if}
