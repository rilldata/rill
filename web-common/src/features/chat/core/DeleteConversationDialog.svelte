<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { deleteConversation } from "./conversation-actions";

  let {
    conversationId,
    open = $bindable(false),
    onDeleted,
  }: {
    conversationId: string;
    open?: boolean;
    onDeleted: () => void;
  } = $props();

  const runtimeClient = useRuntimeClient();

  async function confirmDelete() {
    open = false;

    try {
      await deleteConversation(runtimeClient, conversationId);
      onDeleted();
      eventBus.emit("notification", { message: m.chat_conversation_deleted() });
    } catch (err) {
      console.error("Failed to delete conversation", err);
      eventBus.emit("notification", {
        message: m.chat_conversation_delete_failed(),
        type: "error",
      });
    }
  }
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>{m.chat_conversation_delete_title()}</AlertDialog.Title
      >
      <AlertDialog.Description>
        {m.chat_conversation_delete_description()}
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button type="tertiary" onClick={() => (open = false)}>
        {m.common_cancel()}
      </Button>
      <Button type="destructive" onClick={confirmDelete}>
        {m.common_delete()}
      </Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
