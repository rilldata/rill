<script>
  export let classes = "";
  export let isNull = false;
  export let dark = false;
  $: color = dark ? "text-gray-100" : "text-gray-700";
  $: nullColor = dark ? "text-gray-300" : "text-gray-400";
</script>

<span class="{classes} {isNull ? `${nullColor}` : '{color}'}">
  {#if isNull}
    <span style:font-size=".925em" class="opacity-50 italic">no data</span>
  {:else}
    <slot />
  {/if}
</span>
