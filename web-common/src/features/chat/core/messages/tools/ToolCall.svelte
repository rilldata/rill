<!--
  Renders a tool invocation with collapsible request/response details.
  
  Used in two contexts:
  - "inline": Within ThinkingBlock for tool calls (may be pending)
  - "block": Before output blocks like ChartBlock/FileDiffBlock (always complete)
-->
<script lang="ts">
  import { builderActions, getAttrs } from "bits-ui";
  import CodeBlock from "../../../../../components/code-block/CodeBlock.svelte";
  import * as Collapsible from "../../../../../components/collapsible";
  import CaretDownIcon from "../../../../../components/icons/CaretDownIcon.svelte";
  import LoadingSpinner from "../../../../../components/icons/LoadingSpinner.svelte";
  import type { V1Message, V1Tool } from "../../../../../runtime-client";
  import { MessageContentType } from "../../types";
  import { getToolDisplayName } from "./tool-display-names";
  import { getToolIcon } from "./tool-icons";
  import { isHiddenTool } from "./tool-registry";

  export let message: V1Message;
  export let resultMessage: V1Message | undefined = undefined;
  export let tools: V1Tool[] | undefined = undefined;

  /**
   * Rendering context:
   * - "inline": For thinking block
   * - "block": For output blocks (adds bottom margin before the output)
   */
  export let variant: "inline" | "block" = "inline";

  let isExpanded = false;
  let activeTab: "request" | "response" = "request";

  // Hidden tools are filtered upstream, but check defensively
  $: isHidden = isHiddenTool(message.tool);

  // Result state
  $: hasResult = !!resultMessage;
  $: isError = resultMessage?.contentType === MessageContentType.ERROR;

  // Display name changes based on completion state
  $: toolDisplayName = getToolDisplayName(
    message.tool || "Unknown Tool",
    hasResult,
    tools,
  );

  // Icon based on tool type
  $: ToolIcon = getToolIcon(message.tool);

  // Formatted content
  $: requestContent = formatContent(message);
  $: responseContent = resultMessage ? formatContent(resultMessage) : "";

  function formatContent(msg: V1Message): string {
    const rawContent = msg.contentData || "";

    switch (msg.contentType) {
      case MessageContentType.JSON:
        try {
          return JSON.stringify(JSON.parse(rawContent), null, 2);
        } catch {
          return rawContent;
        }
      case MessageContentType.TEXT:
      case MessageContentType.ERROR:
      default:
        return rawContent;
    }
  }
</script>

{#if !isHidden}
  <Collapsible.Root
    bind:open={isExpanded}
    class="tool-call {variant === 'block' ? 'block' : 'inline'}"
  >
    <Collapsible.Trigger asChild let:builder>
      <button
        class="tool-button"
        {...getAttrs([builder])}
        use:builderActions={{ builders: [builder] }}
      >
        <div class="tool-icon">
          {#if !hasResult && !isExpanded}
            <LoadingSpinner size="14px" />
          {:else if isExpanded}
            <CaretDownIcon size="14" />
          {:else}
            <svelte:component this={ToolIcon} size="14" />
          {/if}
        </div>
        <div class="tool-name">
          {toolDisplayName}
        </div>
      </button>
    </Collapsible.Trigger>

    <Collapsible.Content transition={undefined} class="tool-content">
      <div class="tool-tabs">
        <button
          class="tool-tab"
          class:active={activeTab === "request"}
          on:click={() => (activeTab = "request")}
        >
          Request
        </button>
        <button
          class="tool-tab"
          class:active={activeTab === "response"}
          on:click={() => (activeTab = "response")}
        >
          {isError ? "Error" : "Response"}
        </button>
      </div>

      <div class="tool-tab-content">
        {#if activeTab === "request"}
          <div class="tool-code">
            <CodeBlock code={requestContent} language="json" showCopyButton />
          </div>
        {:else if hasResult}
          <div class="tool-code">
            <CodeBlock code={responseContent} language="json" showCopyButton />
          </div>
        {:else}
          <div class="tool-loading">
            <LoadingSpinner size="0.875rem" />
            <span>Waiting for response...</span>
          </div>
        {/if}
      </div>
    </Collapsible.Content>
  </Collapsible.Root>
{/if}

<style lang="postcss">
  /* Styles for Collapsible.Root - needs :global since it's a child component */
  :global(.tool-call) {
    @apply w-full max-w-full self-start;
  }

  /* Block variant: spacing before the output that follows */
  :global(.tool-call.block) {
    @apply mb-2;
  }

  .tool-button {
    @apply w-full flex items-center gap-1.5 py-1;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-xs text-fg-secondary transition-colors;
  }

  .tool-button:hover {
    @apply text-fg-secondary;
  }

  .tool-icon {
    @apply flex items-center;
  }

  .tool-name {
    @apply font-normal flex-1 text-left;
  }

  /* Styles for Collapsible.Content - needs :global since it's a child component */
  :global(.tool-content) {
    @apply pt-1 ml-5;
  }

  .tool-tabs {
    @apply flex gap-1 pb-2;
  }

  .tool-tab {
    @apply px-2 py-1 text-[0.625rem] font-normal text-fg-secondary;
    @apply bg-transparent border-none cursor-pointer;
    @apply transition-colors rounded;
  }

  .tool-tab:hover {
    @apply text-fg-secondary bg-surface-background;
  }

  .tool-tab.active {
    @apply text-fg-primary bg-gray-100;
  }

  .tool-tab-content {
    @apply bg-surface-background/50 border border-gray-100;
    @apply rounded-md overflow-hidden;
  }

  .tool-code {
    @apply text-[0.625rem];
  }

  .tool-code :global(pre) {
    @apply font-mono text-[0.625rem] leading-snug text-fg-secondary;
    @apply m-0 overflow-x-auto whitespace-pre-wrap break-all;
  }

  .tool-loading {
    @apply p-2 flex items-center gap-2 text-[0.625rem] text-fg-secondary;
  }
</style>
