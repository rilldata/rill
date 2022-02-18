<script>
export let classes = "";
export let isNull = false;
</script>

<span class={isNull ?  'text-gray-400 italic' : `text-gray-700  ${classes}`}>
    {#if isNull} ∅ null
    {:else}
        <slot />
    {/if}
</span>