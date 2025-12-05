<!-- Renders assistant responses from router_agent. -->
<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import {
    getCitationUrlRewriter,
    getMetricsResolverQueryToUrlParamsMapperStore,
  } from "@rilldata/web-common/features/chat/core/messages/rewrite-citation-urls.ts";
  import { derived } from "svelte/store";
  import Markdown from "../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../runtime-client";
  import { extractMessageText } from "../utils";

  export let message: V1Message;

  // Message content and styling
  $: content = extractMessageText(message);

  // Citation URL rewriting for explore dashboards
  // When rendered in an explore context, converts relative citation URLs to full dashboard URLs
  const exploreNameStore = derived(
    page,
    (pageState) => pageState.params.dashboard ?? pageState.params.name ?? "",
  );
  const mapperStore =
    getMetricsResolverQueryToUrlParamsMapperStore(exploreNameStore);

  $: renderedInExplore = !!$exploreNameStore;
  $: convertCitationUrls = renderedInExplore
    ? getCitationUrlRewriter($mapperStore.data)
    : undefined;

  function handleClick(e: MouseEvent) {
    if (!e.target || !(e.target instanceof HTMLElement)) return;
    const urlParams = e.target.getAttribute("data-url-params");
    if (!urlParams) return;
    void goto("?" + urlParams);
  }
</script>

<div class="chat-message">
  <!-- svelte-ignore a11y-no-static-element-interactions a11y-click-events-have-key-events -->
  <div class="chat-message-content" on:click={handleClick}>
    <Markdown {content} converter={convertCitationUrls} />
  </div>
</div>

<style lang="postcss">
  .chat-message {
    @apply max-w-[90%] self-start;
  }

  .chat-message-content {
    @apply px-4 py-2 rounded-2xl;
    @apply text-sm leading-relaxed break-words;
    @apply text-gray-700;
  }
</style>
