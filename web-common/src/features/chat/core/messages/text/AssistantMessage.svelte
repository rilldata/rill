<!-- Renders assistant responses from router_agent. -->
<script lang="ts">
  import { page } from "$app/stores";
  import {
    enhanceCitationLinks,
    getCitationUrlRewriter,
    getMetricsResolverQueryToUrlParamsMapperStore,
  } from "@rilldata/web-common/features/chat/core/messages/text/rewrite-citation-urls.ts";
  import { derived } from "svelte/store";
  import Markdown from "../../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../../runtime-client";
  import { extractMessageText } from "../../utils";

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
</script>

<div class="chat-message">
  <div class="chat-message-content" use:enhanceCitationLinks>
    <Markdown {content} converter={convertCitationUrls} />
  </div>
</div>

<style lang="postcss">
  .chat-message {
    @apply max-w-full;
  }

  .chat-message-content {
    @apply py-2;
    @apply text-sm leading-relaxed break-words;
    @apply text-gray-700;
  }
</style>
