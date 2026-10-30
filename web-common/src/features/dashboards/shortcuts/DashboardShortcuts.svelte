<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Kbd from "@rilldata/web-common/components/Kbd.svelte";
  import {
    DASHBOARD_SHORTCUTS,
    DashboardShortcutAction,
    getDashboardShortcutAction,
    performDashboardShortcut,
  } from "./dashboard-shortcuts";

  let open = false;

  function handleKeydown(event: KeyboardEvent) {
    const action = getDashboardShortcutAction(event);
    if (!action || !performDashboardShortcut(action)) return;

    event.preventDefault();
    event.stopPropagation();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<button
  class="sr-only"
  data-dashboard-shortcut={DashboardShortcutAction.ToggleHelp}
  aria-label="Toggle keyboard shortcuts menu"
  onclick={() => (open = !open)}
>
  Keyboard shortcuts
</button>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-md gap-5">
    <Dialog.Header>
      <Dialog.Title>Keyboard shortcuts</Dialog.Title>
      <Dialog.Description>
        Use these shortcuts while exploring a dashboard.
      </Dialog.Description>
    </Dialog.Header>

    <dl class="grid grid-cols-[1fr_auto] items-center gap-x-6 gap-y-3 text-sm">
      {#each DASHBOARD_SHORTCUTS as shortcut (shortcut.action)}
        <dt class="text-fg-secondary">{shortcut.description}</dt>
        <dd class="flex items-center justify-end gap-1">
          {#each shortcut.keys as key, index (key)}
            {#if index > 0}<span class="text-fg-muted">+</span>{/if}
            <Kbd className="min-w-7 text-center">{key}</Kbd>
          {/each}
        </dd>
      {/each}
    </dl>
  </Dialog.Content>
</Dialog.Root>
