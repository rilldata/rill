<script>
  export let classes = "";
  export let isNull = false;
  export let dark = false;
  $: color = dark ? "" : "text-gray-700";
</script>

<span class="{classes} {color}">
  {#if isNull}
    <span style:font-size=".925em" class="opacity-50 italic">no data</span>
  {:else}
    <slot />
  {/if}
</span>
