<script lang="ts">
  export let classes = "";
  export let isNull = false;
  export let truncate = false;
  export let color = "text-foreground";
</script>

<span class:truncate class="{classes} {color || 'text-foreground'} select-text">
  {#if isNull}
    <span class="text-muted-foreground">-</span>
  {:else}
    <slot />
  {/if}
</span>
