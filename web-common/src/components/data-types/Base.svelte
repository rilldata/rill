<script lang="ts">
  export let classes = "";
  export let isNull = false;
  export let truncate = false;
  export let color = "text-fg-primary";
</script>

<span class:truncate class="{classes} {color || 'text-fg-primary'} select-text">
  {#if isNull}
    <span class="text-fg-secondary">-</span>
  {:else}
    <slot />
  {/if}
</span>
