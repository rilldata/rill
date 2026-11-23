<script lang="ts">
  import { Info } from "lucide-svelte";

  let {
    note,
    detail,
  }: {
    note: string;
    /** The server's own explanation of the cut, written for the model; shown on hover. */
    detail?: string;
  } = $props();
</script>

<p class="banner" role="note" title={detail}>
  <Info size="14" aria-hidden="true" />
  <span>{note}</span>
</p>

<style lang="postcss">
  .banner {
    @apply mx-4 my-1 flex items-start gap-1.5 text-xs leading-4 text-fg-tertiary;
  }
</style>
