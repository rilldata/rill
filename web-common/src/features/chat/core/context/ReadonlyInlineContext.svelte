<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2/index.ts";
  import { builderActions, getAttrs } from "bits-ui";
  import { getInlineChatContextMetadata } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";
  import {
    InlineContextType,
    InlineContextConfig,
    type InlineContext,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

  export let chatContext: InlineContext;

  const contextMetadataStore = getInlineChatContextMetadata();

  $: typeData = chatContext.type
    ? InlineContextConfig[chatContext.type]
    : undefined;
  $: label = typeData?.getLabel(chatContext!, $contextMetadataStore) ?? "";

  $: isMetricsViewContext =
    chatContext.type === InlineContextType.Measure ||
    chatContext.type === InlineContextType.Dimension;
  $: metricsViewName = isMetricsViewContext
    ? InlineContextConfig[InlineContextType.MetricsView]!.getLabel(
        chatContext,
        $contextMetadataStore,
      )
    : "";
</script>

<span class="inline-chat-context">
  {#if metricsViewName}
    <Tooltip.Root>
      <Tooltip.Trigger asChild let:builder>
        <span
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
        >
          {label}
        </span>
      </Tooltip.Trigger>
      <Tooltip.Content>
        From {metricsViewName}
      </Tooltip.Content>
    </Tooltip.Root>
  {:else}
    <span>{label}</span>
  {/if}
</span>

<style lang="postcss">
  .inline-chat-context {
    @apply inline-block gap-1 text-sm underline;
  }
</style>
