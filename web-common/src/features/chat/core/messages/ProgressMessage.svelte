<!--
  Renders AI thinking/reasoning messages that appear during tool execution.
-->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";

  export let message: V1Message;

  $: content = message.contentData || "";
</script>

<div class="progress-message">
  <div class="progress-content">
    {content}
  </div>
</div>

<style lang="postcss">
  .progress-message {
    max-width: 90%;
    align-self: flex-start;
  }

  .progress-content {
    padding: 0.375rem 0.5rem;
    border-radius: 1rem;
    font-size: 0.875rem;
    line-height: 1.5;
    word-break: break-word;
    color: #6b7280;
    font-style: italic;
  }
</style>
