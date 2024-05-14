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
    <span class="text-gray-400">-</span>
  {:else}
    <slot />
  {/if}
</span>
