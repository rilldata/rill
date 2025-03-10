<script lang="ts">
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    type BannerMessage,
    BannerSlot,
  } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";
  import Banner from "./Banner.svelte";

  let banners: (BannerMessage | null)[] = new Array(BannerSlot.Other + 1).fill(
    null,
  );
  const unsubscribe = eventBus.on("banner", (newBanner) => {
    banners[newBanner.slot] = newBanner.message;
  });

  onMount(() => {
    return unsubscribe;
  });
</script>

{#each banners as banner, i (i)}
  {#if banner}
    <Banner {banner} />
  {/if}
{/each}
