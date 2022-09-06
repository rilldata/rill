<script lang="ts">
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import { getContext, hasContext } from "svelte";
  import type { Writable } from "svelte/store";

  export let block = false;
  export let tailwindClasses: string = undefined;
  export let active = false;
  export let level: undefined | "error" = undefined;

  let childRequestedTooltipSuppression;
  let parentHasTooltipSomewhere = hasContext(
    "rill:app:childRequestedTooltipSuppression"
  );
  //
  if (parentHasTooltipSomewhere) {
    childRequestedTooltipSuppression = getContext(
      "rill:app:childRequestedTooltipSuppression"
    ) as Writable<boolean>;
  }

  $: if (parentHasTooltipSomewhere && active) {
    childRequestedTooltipSuppression?.set(true);
  } else {
    childRequestedTooltipSuppression?.set(false);
  }

  $: classes = {
    error: `text-red-600 hover:bg-red-200 ${
      active ? "bg-red-200" : "bg-red-100"
    }`,
    undefined: `text-gray-800 bg-transparent hover:bg-gray-200 ${
      active ? "bg-gray-200" : "bg-transparent"
    }`,
  };
</script>

<button
  class="
{block ? 'flex w-full h-full px-2' : 'inline-flex w-max rounded px-1'} 
  items-center gap-x-2 justify-between 
  {classes[level]}
  {tailwindClasses}"
  on:click
>
  <slot />
  <span class="{active ? '-rotate-180' : ''} transition-transform">
    <CaretDownIcon />
  </span>
</button>
