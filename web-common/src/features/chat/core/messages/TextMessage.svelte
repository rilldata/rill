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

  export let content: string;
  export let role: string;

  const exploreNameStore = derived(
    page,
    (pageState) => pageState.params.dashboard ?? pageState.params.name ?? "",
  );
  $: renderedInExplore = !!$exploreNameStore;

  const mapperStore =
    getMetricsResolverQueryToUrlParamsMapperStore(exploreNameStore);
  $: hasMapper = !!$mapperStore.data;
  $: convertCitationUrls =
    renderedInExplore && hasMapper
      ? getCitationUrlRewriter($mapperStore.data!)
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
    max-width: 90%;
  }

  .chat-message--user {
    align-self: flex-end;
  }

  .chat-message--assistant {
    align-self: flex-start;
  }

  .chat-message-content {
    padding: 0.375rem 0.5rem;
    border-radius: 1rem;
    font-size: 0.875rem;
    line-height: 1.5;
    word-break: break-word;
  }

  .chat-message--user .chat-message-content {
    @apply bg-primary-400 text-white rounded-br-lg;
  }

  .chat-message--assistant .chat-message-content {
    color: #374151;
  }
</style>
