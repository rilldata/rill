<script lang="ts">
  import Markdown from "@rilldata/web-common/components/markdown/Markdown.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1Message } from "@rilldata/web-common/runtime-client";
  import { transformToBlocks } from "../core/messages/block-transform";
  import { extractMessageText } from "../core/utils";

  export let messages: V1Message[];

  $: blocks = transformToBlocks(messages, false, false);

  function messageText(message: V1Message): string {
    return extractMessageText(message).replace(
      /^```markdown\n([\s\S]*)\n```$/,
      "$1",
    );
  }
</script>

<div class="transcript">
  {#each blocks as block (block.id)}
    {#if block.type === "text"}
      {#if block.message.role === "user"}
        <div class="user-message">
          <Markdown content={messageText(block.message)} />
        </div>
      {:else}
        <div class="assistant-message">
          <Markdown content={messageText(block.message)} />
        </div>
      {/if}
    {:else if block.type === "thinking"}
      <details class="thinking">
        <summary>{m.feedback_inbox_transcript_thought()}</summary>
        <ul>
          {#each block.messages as msg (msg.id)}
            <li>
              {#if msg.tool}
                <span class="tool-name">{msg.tool}</span>
              {:else}
                {messageText(msg)}
              {/if}
            </li>
          {/each}
        </ul>
      </details>
    {:else if block.type === "chart"}
      <div class="placeholder">
        {m.feedback_inbox_transcript_chart_placeholder()}
      </div>
    {:else if block.type === "simple-tool-call-block"}
      <div class="placeholder tool-name">{block.message.tool}</div>
    {/if}
  {/each}
</div>

<style lang="postcss">
  .transcript {
    @apply flex flex-col gap-2 py-2;
  }

  .user-message {
    @apply self-end max-w-[85%] px-4 py-2 rounded-2xl rounded-br-lg;
    @apply bg-gray-200 dark:bg-gray-300 text-sm text-fg-primary;
  }

  .assistant-message {
    @apply text-sm leading-relaxed break-words text-fg-primary;
  }

  .thinking {
    @apply text-xs text-fg-muted;
  }

  .thinking summary {
    @apply cursor-pointer select-none;
  }

  .thinking ul {
    @apply pl-4 py-1 flex flex-col gap-0.5 list-disc;
  }

  .tool-name {
    @apply font-mono;
  }

  .placeholder {
    @apply text-xs text-fg-muted italic;
  }
</style>
