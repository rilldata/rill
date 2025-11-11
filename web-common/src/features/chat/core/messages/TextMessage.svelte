<!--
  Renders conversational text exchanges between user and AI assistant.
  Handles router_agent messages (user prompts and assistant responses).
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import { ConversationContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
  import ReadonlyConversationContext from "@rilldata/web-common/features/chat/core/context/ReadonlyConversationContext.svelte";
  import {
    getCitationUrlRewriter,
    getMetricsResolverQueryToUrlParamsMapperStore,
  } from "@rilldata/web-common/features/chat/core/messages/rewrite-citation-urls.ts";
  import { ChevronDownIcon, ChevronRightIcon } from "lucide-svelte";
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
  $: hasMapper = !!$mapperStore.data;
  $: convertCitationUrls =
    renderedInExplore && hasMapper
      ? getCitationUrlRewriter($mapperStore.data!)
      : undefined;

  $: context = ConversationContext.fromMessage(message);
  $: contextRecord = context.record;
  $: hasContext = Object.keys($contextRecord).length > 0;
  let contextOpened = false;
</script>

<div class="chat-message chat-message--{role}">
  <div class="chat-message-content">
    {#if role === "assistant"}
      <Markdown {content} converter={convertCitationUrls} />
    {:else}
      {content}
    {/if}

    {#if hasContext}
      <Collapsible.Root bind:open={contextOpened}>
        <Collapsible.Trigger asChild let:builder>
          <Button
            type="link"
            builders={[builder]}
            class="mt-0.5 text-muted-foreground"
          >
            {#if contextOpened}
              <ChevronDownIcon size="12px" />
            {:else}
              <ChevronRightIcon size="12px" />
            {/if}
            <span class="text-sm text-muted-foreground">
              Additional context
            </span>
          </Button>
        </Collapsible.Trigger>
        <Collapsible.Content class="flex flex-wrap gap-1 items-center">
          <ReadonlyConversationContext {context} />
        </Collapsible.Content>
      </Collapsible.Root>
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
