<script lang="ts">
  import type { ChatContext } from "@rilldata/web-common/features/chat/core/input/types.ts";

  export let chatCtx: ChatContext;

  $: value = chatCtx.values.join(":");
</script>

<span
  class="inline-block items-center gap-1 text-sm underline"
  contenteditable="false"
  data-value={value}
>
  <span>{chatCtx.label}</span>
</span>
