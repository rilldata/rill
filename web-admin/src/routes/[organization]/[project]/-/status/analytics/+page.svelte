<script lang="ts">
  import type { PageData } from "./$types";

  export let data: PageData;
</script>

{#if data.iframeUrl}
  <iframe
    src={data.iframeUrl}
    title="Project Analytics"
    class="iframe"
    frameborder="0"
    allowfullscreen
  />
{:else}
  <p class="text-gray-500">Project analytics are not available.</p>
{/if}

<style lang="postcss">
  .iframe {
    @apply w-full border-0 rounded-sm;
    height: calc(100vh - 200px);
  }
</style>
