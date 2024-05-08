<script>
  export let classes = "";
  export let isNull = false;
  export let dark = false;
  export let truncate = false;
  $: color = dark ? "" : "text-gray-900";
</script>

<span
  class:truncate
  class:inline-block={!truncate}
  class="whitespace-nowrap {classes} {color} break-normal"
>
  {#if isNull}
    <span style:font-size=".925em" class="opacity-50 italic">no data</span>
  {:else}
    <slot />
  {/if}
</span>
