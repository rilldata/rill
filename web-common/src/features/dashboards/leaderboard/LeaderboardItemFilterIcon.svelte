<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import {
    COMPARIONS_COLORS,
    SELECTED_NOT_COMPARED_COLOR,
  } from "@rilldata/web-common/features/dashboards/config";
  export let selectionIndex: number;
  export let excluded = false;
  export let isBeingCompared = false;

  $: selected = selectionIndex >= 0;

  function getColor(i) {
    if (i >= 7) return SELECTED_NOT_COMPARED_COLOR;
    return COMPARIONS_COLORS[i];
  }
</script>

<div style:width="22px" class="grid place-items-center">
  {#if selected && !excluded && isBeingCompared}
    <CheckCircle color={getColor(selectionIndex)} size="18px" />
  {:else if selected && !excluded}
    <Check size="20px" />
  {:else if selected && excluded}
    <Cancel size="20px" />
  {:else}
    <Spacer />
  {/if}
</div>
