<!--
  Lists everything the AI remembers about the current user in this project,
  with per-item edit and delete, "Clear all", "Add memory", and a pause switch.
  Shared by the in-chat dialog (Rill Developer and Rill Cloud) and the Cloud project settings page.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from "@rilldata/web-common/components/dialog";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Textarea from "@rilldata/web-common/components/forms/Textarea.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1AIMemory } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Pencil, Trash2Icon } from "lucide-svelte";
  import {
    createAIMemory,
    deleteAIMemory,
    deleteAllAIMemories,
    MAX_MEMORY_CONTENT_CHARS,
    MEMORY_CATEGORIES,
    MemoryCategory,
    setAIMemoryPaused,
    updateAIMemory,
    useAIMemories,
    useAIMemorySettings,
  } from "./memory-store";

  /** Base path of the chat, used to link a memory to the conversation it was learned from. */
  export let conversationBasePath: string | undefined = undefined;

  const client = useRuntimeClient();
  const memoriesQuery = useAIMemories(client);
  const settingsQuery = useAIMemorySettings(client);

  $: memories = $memoriesQuery.data?.memories ?? [];
  $: enabled = !!$settingsQuery.data?.enabled;
  $: paused = !!$settingsQuery.data?.paused;
  $: maxCount = $settingsQuery.data?.maxCount ?? 0;

  // ----- Edit / add dialog -----
  let editorOpen = false;
  let editing: V1AIMemory | null = null;
  let editorContent = "";
  let editorCategory: string = MemoryCategory.PREFERENCE;
  let saving = false;

  $: editorError =
    editorContent.trim().length > MAX_MEMORY_CONTENT_CHARS
      ? `${editorContent.trim().length} / ${MAX_MEMORY_CONTENT_CHARS}`
      : null;
  $: canSave = !saving && editorContent.trim().length > 0 && !editorError;

  const categoryOptions = MEMORY_CATEGORIES.map((category) => ({
    value: category,
    label: categoryLabel(category),
  }));

  function openAdd() {
    editing = null;
    editorContent = "";
    editorCategory = MemoryCategory.PREFERENCE;
    editorOpen = true;
  }

  function openEdit(memory: V1AIMemory) {
    editing = memory;
    editorContent = memory.content ?? "";
    editorCategory = memory.category ?? MemoryCategory.CONTEXT;
    editorOpen = true;
  }

  async function save() {
    if (!canSave) return;
    saving = true;
    try {
      if (editing?.id) {
        await updateAIMemory(client, editing.id, {
          content: editorContent.trim(),
          category: editorCategory,
        });
      } else {
        await createAIMemory(client, editorCategory, editorContent.trim());
      }
      eventBus.emit("notification", { message: m.chat_memory_saved() });
      editorOpen = false;
    } catch (err) {
      console.error("Failed to save memory", err);
      eventBus.emit("notification", {
        message: m.chat_memory_save_failed(),
        type: "error",
      });
    } finally {
      saving = false;
    }
  }

  // ----- Delete dialogs -----
  let deleting: V1AIMemory | null = null;
  let clearAllOpen = false;

  async function confirmDelete() {
    if (!deleting?.id) return;
    try {
      await deleteAIMemory(client, deleting.id);
      eventBus.emit("notification", { message: m.chat_memory_deleted() });
    } catch (err) {
      console.error("Failed to delete memory", err);
      eventBus.emit("notification", {
        message: m.chat_memory_delete_failed(),
        type: "error",
      });
    } finally {
      deleting = null;
    }
  }

  async function confirmClearAll() {
    try {
      await deleteAllAIMemories(client);
      eventBus.emit("notification", { message: m.chat_memory_cleared() });
    } catch (err) {
      console.error("Failed to clear memories", err);
      eventBus.emit("notification", {
        message: m.chat_memory_delete_failed(),
        type: "error",
      });
    } finally {
      clearAllOpen = false;
    }
  }

  // ----- Pause -----
  async function togglePaused(value: boolean) {
    try {
      await setAIMemoryPaused(client, value);
    } catch (err) {
      console.error("Failed to update memory settings", err);
      eventBus.emit("notification", {
        message: m.chat_memory_save_failed(),
        type: "error",
      });
    }
  }

  function categoryLabel(category: string | undefined): string {
    switch (category) {
      case MemoryCategory.PREFERENCE:
        return m.chat_memory_category_preference();
      case MemoryCategory.DEFINITION:
        return m.chat_memory_category_definition();
      case MemoryCategory.FEEDBACK:
        return m.chat_memory_category_feedback();
      default:
        return m.chat_memory_category_context();
    }
  }

  function formatDate(value: string | undefined): string {
    if (!value) return "";
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }
</script>

<div class="memory-manager" data-testid="memory-manager">
  <p class="memory-description">{m.chat_memory_description()}</p>

  {#if $settingsQuery.isLoading || $memoriesQuery.isLoading}
    <div class="memory-loading">
      <DelayedSpinner isLoading={true} size="20px" />
    </div>
  {:else if !enabled}
    <p class="memory-empty">{m.chat_memory_disabled()}</p>
  {:else}
    <div class="memory-toolbar">
      <label class="memory-pause">
        <Switch
          checked={paused}
          medium
          label={m.chat_memory_pause_label()}
          onCheckedChange={(checked) => void togglePaused(checked)}
        />
        <span class="memory-pause-text">
          <span class="font-medium text-fg-primary"
            >{m.chat_memory_pause_label()}</span
          >
          <span class="text-fg-muted">{m.chat_memory_pause_description()}</span>
        </span>
      </label>
      <div class="memory-toolbar-actions">
        <span class="memory-count">
          {m.chat_memory_count({
            count: String(memories.length),
            max: String(maxCount),
          })}
        </span>
        {#if memories.length > 0}
          <Button type="secondary" onClick={() => (clearAllOpen = true)}>
            {m.chat_memory_clear_all()}
          </Button>
        {/if}
        <Button
          type="primary"
          onClick={openAdd}
          disabled={memories.length >= maxCount}
        >
          {m.chat_memory_add()}
        </Button>
      </div>
    </div>

    {#if paused}
      <p class="memory-paused-notice">{m.chat_memory_paused_notice()}</p>
    {/if}

    {#if memories.length === 0}
      <p class="memory-empty">{m.chat_memory_empty()}</p>
    {:else}
      <div class="memory-table-wrap">
        <table class="memory-table">
          <thead>
            <tr>
              <th>{m.chat_memory_category_label()}</th>
              <th>{m.chat_memory_content_label()}</th>
              <th>{m.chat_memory_learned_from()}</th>
              <th>{m.chat_memory_updated_on()}</th>
              <th><span class="sr-only">{m.chat_memory_edit()}</span></th>
            </tr>
          </thead>
          <tbody>
            {#each memories as memory (memory.id)}
              <tr data-testid="memory-row">
                <td class="memory-category">{categoryLabel(memory.category)}</td
                >
                <td class="memory-content">{memory.content}</td>
                <td class="memory-source">
                  {#if memory.sourceConversationId && conversationBasePath}
                    <a
                      href={`${conversationBasePath}/${memory.sourceConversationId}`}
                      class="text-primary-600 hover:underline"
                    >
                      {m.chat_memory_view_conversation()}
                    </a>
                  {:else if memory.sourceConversationId}
                    <span class="text-fg-muted">
                      {m.chat_memory_view_conversation()}
                    </span>
                  {:else}
                    <span class="text-fg-muted">
                      {m.chat_memory_source_manual()}
                    </span>
                  {/if}
                </td>
                <td class="memory-date">{formatDate(memory.updatedOn)}</td>
                <td class="memory-actions">
                  <DropdownMenu.Root>
                    <DropdownMenu.Trigger class="flex-none">
                      <IconButton rounded ariaLabel={m.chat_memory_edit()}>
                        <ThreeDot size="16px" />
                      </IconButton>
                    </DropdownMenu.Trigger>
                    <DropdownMenu.Content align="end" class="min-w-[120px]">
                      <DropdownMenu.Item
                        class="font-normal flex items-center"
                        onclick={() => openEdit(memory)}
                      >
                        <Pencil size="12px" />
                        <span class="ml-2">{m.chat_memory_edit()}</span>
                      </DropdownMenu.Item>
                      <DropdownMenu.Item
                        class="font-normal flex items-center"
                        type="destructive"
                        onclick={() => (deleting = memory)}
                      >
                        <Trash2Icon size="12px" />
                        <span class="ml-2">{m.chat_memory_delete()}</span>
                      </DropdownMenu.Item>
                    </DropdownMenu.Content>
                  </DropdownMenu.Root>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</div>

<!-- Add / edit -->
<Dialog bind:open={editorOpen}>
  <DialogContent class="max-w-lg">
    <DialogHeader>
      <DialogTitle>
        {editing ? m.chat_memory_edit() : m.chat_memory_add()}
      </DialogTitle>
      <DialogDescription>{m.chat_memory_add_description()}</DialogDescription>
    </DialogHeader>
    <form
      class="memory-form"
      onsubmit={(e) => {
        e.preventDefault();
        void save();
      }}
    >
      <Select
        id="memory-category"
        label={m.chat_memory_category_label()}
        options={categoryOptions}
        bind:value={editorCategory}
      />
      <Textarea
        id="memory-content"
        label={m.chat_memory_content_label()}
        placeholder={m.chat_memory_content_placeholder()}
        rows={3}
        errors={editorError}
        bind:value={editorContent}
      />
    </form>
    <DialogFooter>
      <Button type="tertiary" onClick={() => (editorOpen = false)}>
        {m.common_cancel()}
      </Button>
      <Button type="primary" disabled={!canSave} onClick={save}>
        {m.chat_memory_save()}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>

<!-- Delete one -->
<AlertDialog.Root
  open={deleting !== null}
  onOpenChange={(open) => {
    if (!open) deleting = null;
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>{m.chat_memory_delete_title()}</AlertDialog.Title>
      <AlertDialog.Description>
        {m.chat_memory_delete_description()}
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button type="tertiary" onClick={() => (deleting = null)}>
        {m.common_cancel()}
      </Button>
      <Button type="destructive" onClick={confirmDelete}>
        {m.chat_memory_delete()}
      </Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<!-- Clear all -->
<AlertDialog.Root bind:open={clearAllOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>{m.chat_memory_clear_all_title()}</AlertDialog.Title>
      <AlertDialog.Description>
        {m.chat_memory_clear_all_description()}
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button type="tertiary" onClick={() => (clearAllOpen = false)}>
        {m.common_cancel()}
      </Button>
      <Button type="destructive" onClick={confirmClearAll}>
        {m.chat_memory_clear_all()}
      </Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<style lang="postcss">
  .memory-manager {
    @apply flex flex-col gap-y-4 w-full min-w-0;
  }

  .memory-description {
    @apply text-sm text-fg-secondary;
  }

  .memory-loading {
    @apply flex justify-center py-6;
  }

  .memory-empty {
    @apply text-sm text-fg-muted py-6 text-center;
  }

  .memory-toolbar {
    @apply flex flex-wrap items-center justify-between gap-3;
  }

  .memory-pause {
    @apply flex items-center gap-x-3 cursor-pointer;
  }

  .memory-pause-text {
    @apply flex flex-col text-xs;
  }

  .memory-toolbar-actions {
    @apply flex items-center gap-x-2;
  }

  .memory-count {
    @apply text-xs text-fg-muted tabular-nums;
  }

  .memory-paused-notice {
    @apply text-xs text-fg-secondary rounded-md bg-surface-subtle px-3 py-2;
  }

  .memory-table-wrap {
    @apply w-full overflow-x-auto rounded-md border border-border;
  }

  .memory-table {
    @apply w-full border-collapse text-sm;
  }

  .memory-table th {
    @apply text-left text-xs font-medium text-fg-muted uppercase tracking-wide;
    @apply px-3 py-2 bg-surface-subtle border-b border-border whitespace-nowrap;
  }

  .memory-table td {
    @apply px-3 py-2 border-b border-border align-top;
  }

  .memory-table tr:last-child td {
    @apply border-b-0;
  }

  .memory-category {
    @apply whitespace-nowrap text-fg-secondary;
  }

  .memory-content {
    @apply text-fg-primary;
    min-width: 16rem;
  }

  .memory-source,
  .memory-date {
    @apply whitespace-nowrap text-xs text-fg-secondary;
  }

  .memory-actions {
    @apply w-10;
  }

  .memory-form {
    @apply flex flex-col gap-y-3 py-2;
  }
</style>
