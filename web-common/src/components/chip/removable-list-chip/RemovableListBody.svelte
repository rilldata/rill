<script lang="ts">
  export let label: string;
  export let values: string[];
  export let show = 1;
  export let labelMaxWidth = "160px";
  export let valueMaxWidth = "320px";

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
</div>
