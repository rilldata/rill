<script lang="ts">
  export let paddingLeft: number;
  export let paddingRight: number;
  export let element = "td";
</script>

<tr>
  {#if paddingLeft > 0}
    <svelte:element this={element} style={`width: ${paddingLeft}px;`} />
  {/if}
  <slot />
  {#if paddingRight > 0}
    <svelte:element this={element} style={`width: ${paddingRight}px;`} />
  {/if}
</tr>
