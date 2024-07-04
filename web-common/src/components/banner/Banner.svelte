<script lang="ts">
  import { BannerMessage } from "../../lib/event-bus/events";
  import AlertCircleIcon from "../icons/AlertCircleOutline.svelte";

  export let banner: BannerMessage;
</script>

<header class="app-banner">
  <div class="banner-content">
    <AlertCircleIcon size="16px" className="text-secondary-700" />
    {#if banner.includesHtml}
      <p class="banner-message">{@html banner.message}</p>
    {:else}
      <p class="banner-message">{banner.message}</p>
    {/if}
  </div>
</header>

<style lang="postcss">
  .app-banner {
    @apply h-7 bg-secondary-100 border-b;
  }

  .banner-content {
    @apply px-2 py-1.5;
    @apply flex justify-center items-center gap-x-2;
  }

  .banner-message {
    @apply text-xs text-secondary-900 font-medium;
  }

  .banner-message :global(a) {
    @apply text-primary-600;
  }

  .banner-message :global(a:hover) {
    @apply text-primary-700;
  }
</style>
