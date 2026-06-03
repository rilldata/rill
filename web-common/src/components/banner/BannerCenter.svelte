<script lang="ts">
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { type BannerEvent } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";
  import Banner from "./Banner.svelte";
  import {
    dismissBanner,
    isBannerDismissed,
  } from "@rilldata/web-common/components/banner/banner-dismiss.ts";

  let banners: BannerEvent[] = [];

  const unsubscribeAddBanner = eventBus.on("add-banner", (newBanner) => {
    const existingIdx = banners.findIndex((b) => b.id === newBanner.id);
    if (existingIdx === -1) {
      banners.push(newBanner);
    } else if (existingIdx >= 0) {
      banners[existingIdx] = newBanner;
    }
    banners = banners.sort((a, b) => a.priority - b.priority);
  });

  const unsubscribeRemoveBanner = eventBus.on("remove-banner", removeBanner);

  onMount(() => {
    return () => {
      unsubscribeAddBanner();
      unsubscribeRemoveBanner();
    };
  });

  function removeBanner(bannerId: string) {
    banners = banners.filter((banner) => banner.id !== bannerId);
  }

  function handleDismiss(bannerEvent: BannerEvent) {
    dismissBanner(bannerEvent.message.dismissible);
    removeBanner(bannerEvent.id);
  }
</script>

{#each banners as bannerEvent (bannerEvent.id)}
  {#if bannerEvent.message && !isBannerDismissed(bannerEvent.message.dismissible)}
    <Banner
      banner={bannerEvent.message}
      dismiss={() => handleDismiss(bannerEvent)}
    />
  {/if}
{/each}
