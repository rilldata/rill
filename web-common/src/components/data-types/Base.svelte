<script>
  export let classes = "";
  export let isNull = false;
  export let dark = false;
  $: color = dark ? "" : "text-gray-900";
</script>

<span class=" whitespace-nowrap inline-block {classes} {color} break-normal">
  {#if isNull}
    <span style:font-size=".925em" class="opacity-50 italic">no data</span>
  {:else}
    <slot />
  {/if}
</span>
