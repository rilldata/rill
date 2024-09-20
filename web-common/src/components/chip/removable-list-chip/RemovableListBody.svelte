<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let label: string;
  export let values: string[];
  export let show = 1;
  export let labelMaxWidth = "160px";
  export let valueMaxWidth = "320px";
  export let active = false;
  export let readOnly = false;

  $: whatsLeft = values.length - show;
</script>

<div class="flex gap-x-2 items-center">
  <span class="font-bold truncate" style:max-width={labelMaxWidth}>
    {label}
  </span>

  {#each values.slice(0, show) as value (value)}
    <span class="truncate" style:max-width={valueMaxWidth}>
      {value}
    </span>
  {/each}

  {#if values.length > 1}
    <span class="italic">
      +{whatsLeft} other{#if whatsLeft !== 1}s{/if}
    </span>
  {/if}
  {#if !readOnly}
    <div class="transition-transform -mr-1" class:-rotate-180={active}>
      <CaretDownIcon size="10px" />
    </div>
  {/if}
</div>
