<!-- Conversation title that opens a menu with actions for the conversation. -->
<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import DeleteConversationDialog from "@rilldata/web-common/features/chat/core/DeleteConversationDialog.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ChevronDown, Pencil, Trash2Icon } from "lucide-svelte";

  let {
    conversationId,
    title,
    onRename,
    onDeleted,
  }: {
    conversationId: string;
    title: string;
    onRename: () => void;
    onDeleted: () => void;
  } = $props();

  let confirmingDelete = $state(false);
  // Set when an action is picked, so that closing the menu does not move focus back to the trigger,
  // away from the title editor or the delete dialog.
  let actionPicked = false;

  function pickAction(action: () => void) {
    actionPicked = true;
    action();
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger
    class="flex items-center gap-x-1 min-w-0 px-1.5 py-0.5 -ml-1.5 rounded-md text-sm font-semibold text-fg-secondary hover:bg-surface-muted data-[state=open]:bg-surface-muted"
  >
    <span class="truncate">{title}</span>
    <ChevronDown size="14px" class="shrink-0 text-fg-muted" />
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="start"
    class="min-w-[160px]"
    onCloseAutoFocus={(e: Event) => {
      if (actionPicked) e.preventDefault();
      actionPicked = false;
    }}
  >
    <DropdownMenu.Item onclick={() => pickAction(onRename)}>
      <Pencil size="12px" />
      {m.chat_conversation_rename()}
    </DropdownMenu.Item>
    <DropdownMenu.Item
      type="destructive"
      onclick={() => pickAction(() => (confirmingDelete = true))}
    >
      <Trash2Icon size="12px" />
      {m.chat_conversation_delete()}
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<DeleteConversationDialog
  {conversationId}
  bind:open={confirmingDelete}
  {onDeleted}
/>
