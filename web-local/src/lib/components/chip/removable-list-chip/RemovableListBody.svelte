<!-- @component 
  renders the body content of a filter set chip:
  - a label for the current dimension
  - a certain number of "show" values (default 1)
  - an indication of how many other dimensions are selected past the show values
-->
<script lang="ts">
  export let label: string;
  export let values: string[];
  export let show = 1;
  export let labelMaxWidth = "160px";
  export let valueMaxWidth = "320px";

  $: visibleValues = values.slice(0, show);
  $: whatsLeft = values.length - show;
</script>

<div class="flex gap-x-2">
  <div
    class="font-bold text-ellipsis overflow-hidden whitespace-nowrap"
    style:max-width={labelMaxWidth}
  >
    {label}
  </div>
  <div class="flex flex-wrap gap-x-2 gap-y-1">
    {#each visibleValues as value}
      <div
        class="text-ellipsis overflow-hidden whitespace-nowrap"
        style:max-width={valueMaxWidth}
      >
        {value}
      </div>
    {/each}
    {#if values.length > 1}
      <div class="italic">
        + {whatsLeft} other{#if whatsLeft !== 1}s{/if}
      </div>
    {/if}
  </div>
</div>
