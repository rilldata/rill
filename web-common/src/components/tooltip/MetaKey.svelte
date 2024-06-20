<script lang="ts">
  export let action = "Click";
  export let withAlt = false;

  function isMac() {
    return window.navigator.userAgent.includes("Macintosh");
  }
</script>

{#if withAlt}
  {#if isMac()}
    <span style="font-family: var(--system);">⌥</span>
  {:else}alt{/if} +
{/if}
{#if isMac()}
  <span style="font-family: var(--system);">⌘</span>
{:else}ctrl{/if} + {action}
