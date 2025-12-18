<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { getNextLimitLabel } from "./pivot-constants";

  export let currentLimit: number;
  export let dimensionLabel: string;
  export let depth: number = 0;
  export let assembled = true;

  $: nextLimitLabel = getNextLimitLabel(currentLimit);
</script>

<div class="show-more-cell" style:padding-left="{depth * 14}px">
  <Spacer size="16px" />
  <span class="italic {assembled ? 'ui-copy' : 'ui-copy-inactive'}">
    Increase limit to {nextLimitLabel} on '{dimensionLabel}'
  </span>
</div>

<style lang="postcss">
  .show-more-cell {
    @apply flex items-center gap-x-0.5 h-full;
  }

  .show-more-cell:hover span {
    @apply text-primary-600;
  }
</style>
