<script lang="ts">
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useIsSourceNotSaved } from "../selectors";
  import { useSourceStore } from "../sources-store";

  export let sourceName: string;

  $: sourceStore = useSourceStore(sourceName);

  // Include `clientYAML` in the reactive statement to recompute the value whenever `clientYAML` changes
  $: isSourceNotSaved =
    $sourceStore.clientYAML &&
    useIsSourceNotSaved($runtime.instanceId, sourceName);
</script>

{#if isSourceNotSaved}
  <div>NOT SAVED</div>
{/if}
