<!--
  Renders conversational text exchanges between user and AI assistant.
  Handles router_agent messages (user prompts and assistant responses).
-->
<script lang="ts">
  import { page } from "$app/stores";
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
  $: role = message.role || "assistant";
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

<div class="chat-message chat-message--{role}">
  <div class="chat-message-content">
    {#if role === "assistant"}
      <Markdown {content} converter={convertCitationUrls} />
    {:else}
      {content}
    {/if}
  </div>
</div>

<style lang="postcss">
  .chat-message {
    @apply max-w-[90%];
  }

  .chat-message--user {
    @apply self-end;
  }

  .chat-message--assistant {
    @apply self-start;
  }

  .chat-message-content {
    @apply px-2 py-1.5 rounded-2xl text-sm leading-relaxed break-words;
  }

  .chat-message--user .chat-message-content {
    @apply bg-primary-400 text-white rounded-br-lg;
  }

  .chat-message--assistant .chat-message-content {
    @apply text-gray-700;
  }
</style>
