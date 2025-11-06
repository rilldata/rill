<!--
  Renders AI thinking/reasoning messages that appear during tool execution.
-->
<script lang="ts">
  import Markdown from "../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../runtime-client";

  export let message: V1Message;

  $: content = message.contentData || "";
</script>

<div class="progress-message">
  <div class="progress-content">
    <Markdown {content} />
  </div>
</div>

<style lang="postcss">
  .progress-message {
    @apply max-w-[90%] self-start;
  }

  .progress-content {
    @apply px-2 py-1.5 rounded-2xl text-sm leading-relaxed break-words text-gray-500 italic;
  }
</style>
