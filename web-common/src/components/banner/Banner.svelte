<script lang="ts">
  import { BannerMessage } from "../../lib/event-bus/events";
  import AlertCircleIcon from "../icons/AlertCircleOutline.svelte";
  import CheckCircleOutline from "../icons/CheckCircleOutline.svelte";
  import MoonCircleOutline from "../icons/MoonCircleOutline.svelte";

  export let banner: BannerMessage;

  $: ctaElement = banner.ctaUrl ? "a" : "span";
  $: ctaProps = {
    ...(banner.ctaUrl ? { href: banner.ctaUrl } : {}),
    ...(banner.ctaTarget ? { target: banner.ctaTarget } : {}),
    ...(banner.ctaCallback ? { "on:click": banner.ctaCallback } : {}),
  };

  const IconMap = {
    alert: AlertCircleIcon,
    check: CheckCircleOutline,
    sleep: MoonCircleOutline,
    // TODO
    // loading: AlertCircleIcon,
  };
</script>

<header class="{banner.type} app-banner">
  <div class="banner-content">
    {#if banner.iconType in IconMap}
      <svelte:component
        this={IconMap[banner.iconType]}
        size="12px"
        class="banner-icon"
      />
    {/if}
    {#if banner.includesHtml}
      <p class="banner-message">{@html banner.message}</p>
    {:else}
      <p class="banner-message">{banner.message}</p>
    {/if}
    {#if banner.ctaText}
      <svelte:element this={ctaElement} class="banner-cta" {...ctaProps}>
        {banner.ctaText}
      </svelte:element>
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

  .banner-icon {
    @apply text-secondary-900;
  }

  .banner-message {
    @apply text-xs text-secondary-900 font-medium;
  }

  .banner-message :global(a) {
    @apply text-primary-600;
  }

  .banner-cta {
    @apply text-primary-600 cursor-pointer;
  }

  .banner-message :global(a:hover) {
    @apply text-primary-700;
  }

  .success.app-banner {
    @apply bg-primary-100;
  }
  .success .banner-message,
  .success .banner-icon {
    @apply text-primary-800;
  }

  .info.app-banner {
    @apply bg-slate-100;
  }
  .info .banner-message,
  .info .banner-icon {
    @apply text-slate-700;
  }

  .warning.app-banner {
    @apply bg-yellow-100;
  }
  .warning .banner-message,
  .warning .banner-icon {
    @apply text-yellow-700;
  }

  .error.app-banner {
    @apply bg-red-100;
  }
  .error .banner-message,
  .error .banner-icon {
    @apply text-red-700;
  }
</style>
