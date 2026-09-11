<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ArrowUp } from "lucide-svelte";
</script>

<!-- Stand-in for ChatInput for visitors who can't send messages (e.g. anonymous recipients of a shared conversation). -->
<div class="read-only-chat-input" aria-disabled="true">
  <p class="text-sm text-fg-muted">{m.chat_read_only_notice()}</p>
  <div class="flex flex-row">
    <div class="grow"></div>
    <Button type="primary" label={m.chat_send_message()} disabled square>
      <ArrowUp size="16px" />
    </Button>
  </div>
</div>

<style lang="postcss">
  .read-only-chat-input {
    @apply flex flex-col gap-1 p-3;
    @apply border rounded-md bg-surface-subtle cursor-not-allowed;
  }
</style>
