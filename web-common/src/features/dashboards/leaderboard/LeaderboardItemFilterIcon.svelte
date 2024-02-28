<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";

  export let selectionIndex: number;
  export let excluded = false;
  export let isBeingCompared = false;
  export let color: string;

  $: selected = selectionIndex >= 0;
</script>

<div style:width="22px" class="grid place-items-center">
  {#if selected && !excluded && isBeingCompared}
    <CheckCircle className="fill-{color}" size="18px" />
  {:else if selected && !excluded}
    <Check size="20px" />
  {:else if selected && excluded}
    <Cancel size="20px" />
  {:else}
    <Spacer />
  {/if}
</div>
