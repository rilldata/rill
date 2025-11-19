<!--
  Renders conversational text exchanges between user and AI assistant.
  Handles router_agent messages (user prompts and assistant responses).
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { convertContextToHtml } from "@rilldata/web-common/features/chat/core/context/conversions.ts";
  import { getContextMetadataStore } from "@rilldata/web-common/features/chat/core/context/get-context-metadata-store.ts";
  import {
    getCitationUrlRewriter,
    getMetricsResolverQueryToUrlParamsMapperStore,
  } from "@rilldata/web-common/features/chat/core/messages/rewrite-citation-urls.ts";
  import { derived } from "svelte/store";
  import Markdown from "../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../runtime-client";
  import { extractMessageText } from "../utils";
  import DOMPurify from "dompurify";

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

  const contextMetadataStore = getContextMetadataStore();
</script>

<div class="chat-message chat-message--{role}">
  <div class="chat-message-content">
    {#if role === "assistant"}
      <Markdown {content} converter={convertCitationUrls} />
    {:else}
      {@html DOMPurify.sanitize(
        convertContextToHtml(content, $contextMetadataStore),
      )}
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
    @apply px-4 py-2 rounded-2xl;
    @apply text-sm leading-relaxed break-words;
  }

  .chat-message--user .chat-message-content {
    @apply bg-primary-100/50 text-foreground rounded-br-lg;
  }

  .chat-message--assistant .chat-message-content {
    @apply text-gray-700;
  }
</style>
