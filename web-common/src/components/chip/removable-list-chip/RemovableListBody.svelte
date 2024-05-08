<!-- @component 
  renders the body content of a filter set chip:
  - a label for the current dimension
  - a certain number of "show" values (default 1)
  - an indication of how many other dimensions are selected past the show values
-->
<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  export let label: string;
  export let values: string[];
  export let show = 1;
  export let labelMaxWidth = "160px";
  export let valueMaxWidth = "320px";
  export let active = false;

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
  <div class="flex flex-wrap flex-row items-center gap-y-1 gap-x-2">
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
        +{whatsLeft} other{#if whatsLeft !== 1}s{/if}
      </div>
    {/if}
    <IconSpaceFixer pullRight>
      <div class="transition-transform" class:-rotate-180={active}>
        <CaretDownIcon size="10px" />
      </div>
    </IconSpaceFixer>
  </div>
</div>
