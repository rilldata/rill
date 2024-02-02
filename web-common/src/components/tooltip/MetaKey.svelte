<script lang="ts">
  function isMac() {
    return window.navigator.userAgent.includes("Macintosh");
  }
</script>

{#if isMac()}
  <span style="font-family: var(--system);">⌘</span>
{:else}ctrl{/if} + Click
