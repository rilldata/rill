<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import CaretDownFilledIcon from "@rilldata/web-common/components/icons/CaretDownFilledIcon.svelte";
  import CaretRightFilledIcon from "@rilldata/web-common/components/icons/CaretRightFilledIcon.svelte";
  import { ConversationContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
  import ConversationContextEntryDisplay from "@rilldata/web-common/features/chat/core/context/ConversationContextEntryDisplay.svelte";
  import {
    getCitationUrlRewriter,
    getMetricsResolverQueryToUrlParamsMapperStore,
  } from "@rilldata/web-common/features/chat/core/messages/rewrite-citation-urls.ts";
  import { derived } from "svelte/store";
  import Markdown from "../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../runtime-client";

  export let message: V1Message;
  export let content: string;

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

  const context = new ConversationContext();
  $: context.parseContext(message);
  $: contextData = context.data;
  $: hasContext = $contextData.length > 0;
  let contextOpened = false;

  $: role = message.role;
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
          <Button type="link" builders={[builder]} class="ml-1">
            {#if contextOpened}
              <CaretDownFilledIcon size="12px" fillColor="white" />
            {:else}
              <CaretRightFilledIcon size="12px" fillColor="white" />
            {/if}
            <span class="text-white">Additional context</span>
          </Button>
        </Collapsible.Trigger>
        <Collapsible.Content class="flex flex-wrap gap-1 items-center">
          {#each $contextData as entry (entry.type)}
            <ConversationContextEntryDisplay {context} {entry} />
          {/each}
        </Collapsible.Content>
      </Collapsible.Root>
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
