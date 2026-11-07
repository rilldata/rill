<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Kbd from "@rilldata/web-common/components/Kbd.svelte";
  import {
    getDashboardShortcuts,
    type DashboardShortcut,
    type DashboardShortcutView,
    DASHBOARD_SHORTCUT_VIEWS,
    DashboardShortcutAction,
    getDashboardShortcutAction,
    performDashboardShortcut,
  } from "./dashboard-shortcuts";

  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { measureSelection } from "../time-series/measure-selection/measure-selection";

  export let view: DashboardShortcutView = "explore";

  const explainEnabled = measureSelection.getEnabledStore();
  let open = false;
  let shortcuts: DashboardShortcut[] = [];

  function viewLabel(view: DashboardShortcutView) {
    switch (view) {
      case "canvas":
        return m.dashboard_shortcuts_canvas();
      case "explore":
        return m.dashboard_explore();
      case "pivot":
        return m.dashboard_pivot();
    }
  }

  $: groups = [
    {
      id: "shared",
      title: m.dashboard_shortcuts_shared(),
      items: shortcuts.filter(
        (shortcut) => shortcut.views.length === DASHBOARD_SHORTCUT_VIEWS.length,
      ),
    },
    {
      id: "view",
      title: viewLabel(view),
      items: shortcuts.filter(
        (shortcut) => shortcut.views.length !== DASHBOARD_SHORTCUT_VIEWS.length,
      ),
    },
  ];

  function toggleHelp() {
    if (!open)
      shortcuts = getDashboardShortcuts({
        explainEnabled: $explainEnabled,
        view,
      });
    open = !open;
  }

  function handleKeydown(event: KeyboardEvent) {
    const action = getDashboardShortcutAction(event, { helpOpen: open });
    if (!action || !performDashboardShortcut(action)) return;

    event.preventDefault();
    event.stopPropagation();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<button
  class="sr-only"
  data-dashboard-shortcut={DashboardShortcutAction.ToggleHelp}
  aria-label={m.dashboard_shortcuts_toggle_aria()}
  onclick={toggleHelp}
>
  {m.dashboard_shortcuts_title()}
</button>

<Dialog.Root bind:open>
  <Dialog.Content
    class="max-w-md max-h-[calc(100dvh-2rem)] overflow-y-auto gap-5"
  >
    <Dialog.Header>
      <Dialog.Title>{m.dashboard_shortcuts_title()}</Dialog.Title>
      <Dialog.Description>
        {m.dashboard_shortcuts_description()}
      </Dialog.Description>
    </Dialog.Header>

    <div
      class="flex flex-wrap gap-2"
      aria-label={m.dashboard_shortcuts_views_aria()}
    >
      {#each DASHBOARD_SHORTCUT_VIEWS as candidate (candidate)}
        <span
          class="rounded-full px-3 py-1 text-xs border"
          class:bg-surface-subtle={candidate === view}
          class:font-semibold={candidate === view}
          class:text-fg-muted={candidate !== view}
          aria-current={candidate === view ? "true" : undefined}
          >{viewLabel(candidate)}</span
        >
      {/each}
    </div>

    {#each groups as group (group.id)}
      {#if group.items.length}
        <section aria-label={group.title} class="space-y-3">
          <h3 class="text-xs font-semibold uppercase text-fg-muted">
            {group.title}
          </h3>
          <dl
            class="grid grid-cols-[1fr_auto] items-center gap-x-6 gap-y-3 text-sm"
          >
            {#each group.items as shortcut (shortcut.id)}
              <dt class="text-fg-secondary">
                {shortcut.description}
                {#if shortcut.context}
                  <span class="block mt-1 text-xs text-fg-muted"
                    >{shortcut.context}</span
                  >
                {/if}
              </dt>
              <dd class="flex items-center justify-end gap-1">
                {#each shortcut.keys as key, index (key)}
                  {#if index > 0}<span class="text-fg-muted">+</span>{/if}
                  <Kbd className="min-w-7 text-center">{key}</Kbd>
                {/each}
              </dd>
            {/each}
          </dl>
        </section>
      {/if}
    {/each}
  </Dialog.Content>
</Dialog.Root>
