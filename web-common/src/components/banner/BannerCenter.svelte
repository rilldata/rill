<script lang="ts">
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";
  import Banner from "./Banner.svelte";

  let banner: BannerMessage | null = null;

  onMount(() => {
    const unsubscribe = eventBus.on("banner", (newBanner) => {
      banner = newBanner;
    });

    return unsubscribe;
  });
</script>

{#if banner}
  <Banner {banner} />
{/if}
