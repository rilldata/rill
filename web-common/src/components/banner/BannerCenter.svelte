<script lang="ts">
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { type BannerEvent } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";
  import Banner from "./Banner.svelte";

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

  const unsubscribeRemoveBanner = eventBus.on("remove-banner", (bannerId) => {
    banners = banners.filter((banner) => banner.id !== bannerId);
  });

  onMount(() => {
    return () => {
      unsubscribeAddBanner();
      unsubscribeRemoveBanner();
    };
  });
</script>

{#each banners as { message, id } (id)}
  {#if message}
    <Banner banner={message} />
  {/if}
{/each}
