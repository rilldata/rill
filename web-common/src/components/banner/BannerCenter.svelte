<script lang="ts">
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { type BannerEvent } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";
  import Banner from "./Banner.svelte";

  let banners: BannerEvent[] = [];
  const unsubscribe = eventBus.on("banner", (newBanner) => {
    const existingIdx = banners.findIndex((b) => b.id === newBanner.id);
    if (existingIdx === -1 && newBanner.message) {
      console.log("New banner", newBanner);
      banners.push(newBanner);
    } else if (existingIdx >= 0) {
      console.log("Replace banner", newBanner);
      if (newBanner.message) {
        banners[existingIdx] = newBanner;
      } else {
        banners.splice(existingIdx, 1);
      }
    }
    banners = banners.sort((a, b) => a.priority - b.priority);
  });

  $: console.log(banners);

  onMount(() => {
    return unsubscribe;
  });
</script>

{#each banners as { message, id } (id)}
  {#if message}
    <Banner banner={message} />
  {/if}
{/each}
