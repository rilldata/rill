<!--
  Renders AI thinking/reasoning messages that appear during tool execution.
-->
<script lang="ts">
  import CaretDownIcon from "../../../../components/icons/CaretDownIcon.svelte";
  import ChevronRight from "../../../../components/icons/ChevronRight.svelte";
  import Markdown from "../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../runtime-client";

  export let message: V1Message;

  let isExpanded = true;

  $: content = message.contentData || "";

  function toggleExpanded() {
    isExpanded = !isExpanded;
  }
</script>

<div class="progress-message">
  <button class="progress-header" on:click={toggleExpanded}>
    <div class="progress-icon">
      {#if isExpanded}
        <CaretDownIcon size="14" />
      {:else}
        <ChevronRight size="14" />
      {/if}
    </div>
    <div class="progress-title">Thinking...</div>
  </button>

  {#if isExpanded && content}
    <div class="progress-content">
      <Markdown {content} />
    </div>
  {/if}
</div>

<style lang="postcss">
  .progress-message {
    @apply w-full max-w-[90%] self-start;
  }

  .progress-header {
    @apply w-full flex items-center gap-1.5 px-1 py-1;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-xs text-gray-400 transition-colors;
  }

  .progress-header:hover {
    @apply text-gray-500;
  }

  .progress-icon {
    @apply flex items-center text-gray-400;
  }

  .progress-title {
    @apply flex-1 text-left font-normal;
  }

  .progress-content {
    @apply px-6 py-1 text-xs text-gray-400 leading-relaxed break-words;
  }
</style>
