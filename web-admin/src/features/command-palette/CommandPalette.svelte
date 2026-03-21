<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { browser } from "$app/environment";
  import { Command } from "cmdk-sv";
  import * as Dialog from "@rilldata/web-common/components/dialog/index.js";
  import { createAdminServiceListProjectsForOrganization } from "@rilldata/web-admin/client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { searchIndex } from "./search-orchestrator";
  import { buildRoute } from "./route-builders";
  import { prefetchAllResources } from "./resource-prefetch";
  import CommandPaletteItem from "./CommandPaletteItem.svelte";
  import type { SearchableItem } from "./types";

  let open = false;
  let query = "";

  $: orgName = $page.params.organization;
  $: isMac = browser && window.navigator.userAgent.includes("Macintosh");

  $: projectListQuery = createAdminServiceListProjectsForOrganization(
    orgName,
    undefined,
    {
      query: {
        enabled: !!orgName,
        staleTime: 5 * 60 * 1000,
      },
    },
  );

  $: projectItems = buildProjectItems(orgName, $projectListQuery.data?.projects);

  let resourceItems: SearchableItem[] = [];
  const queryClient = useQueryClient();

  $: if ($projectListQuery.data?.projects) {
    const names = $projectListQuery.data.projects
      .filter((p): p is { name: string } => !!p.name)
      .map((p) => p.name);
    void prefetchAllResources(queryClient, orgName, names, (items) => {
      resourceItems = items;
    });
  }

  $: if (!open) {
    query = "";
  }

  $: searchItems = [...projectItems, ...resourceItems];
  $: results = searchIndex(searchItems, query);
  $: hasResults =
    results.projects.length > 0 ||
    results.dashboards.length > 0 ||
    results.reports.length > 0 ||
    results.alerts.length > 0;
  $: isLoading = $projectListQuery.isLoading;
  $: isError = $projectListQuery.isError;

  function buildProjectItems(
    org: string,
    projects: Array<{ name?: string }> | undefined,
  ): SearchableItem[] {
    if (!projects) return [];
    return projects
      .filter((p) => p.name)
      .map((p) => ({
        name: p.name!,
        type: "project" as const,
        projectName: p.name!,
        orgName: org,
        route: buildRoute("project", org, p.name!, p.name!),
      }));
  }

  function handleSelect(item: SearchableItem) {
    open = false;
    query = "";
    void goto(item.route);
  }

  function handleKeydown(e: KeyboardEvent) {
    const metaOrCtrl = isMac ? e.metaKey : e.ctrlKey;
    if (metaOrCtrl && e.key === "k") {
      e.preventDefault();
      open = !open;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<Dialog.Root bind:open>
  <Dialog.Content class="command-palette-dialog" noClose>
    <div class="command-palette">
      <Command.Root shouldFilter={false}>
        <Command.Input
          autofocus
          placeholder="Search projects, dashboards, reports..."
          bind:value={query}
        />
        <div class="command-palette-separator" />
        <Command.List>
          {#if isLoading}
            <div class="command-palette-status">Loading...</div>
          {:else if isError}
            <div class="command-palette-status">Unable to load search data</div>
          {:else if query.length < 2}
            <div class="command-palette-status">Type to search...</div>
          {:else if !hasResults}
            <Command.Empty>No results found.</Command.Empty>
          {:else}
            {#if results.projects.length > 0}
              <Command.Group heading="Projects">
                {#each results.projects as item (item.route)}
                  <Command.Item
                    value={item.route}
                    onSelect={() => handleSelect(item)}
                  >
                    <CommandPaletteItem {item} />
                  </Command.Item>
                {/each}
              </Command.Group>
            {/if}

            {#if results.dashboards.length > 0}
              <Command.Group heading="Dashboards">
                {#each results.dashboards as item (item.route)}
                  <Command.Item
                    value={item.route}
                    onSelect={() => handleSelect(item)}
                  >
                    <CommandPaletteItem {item} />
                  </Command.Item>
                {/each}
              </Command.Group>
            {/if}

            {#if results.reports.length > 0}
              <Command.Group heading="Reports">
                {#each results.reports as item (item.route)}
                  <Command.Item
                    value={item.route}
                    onSelect={() => handleSelect(item)}
                  >
                    <CommandPaletteItem {item} />
                  </Command.Item>
                {/each}
              </Command.Group>
            {/if}

            {#if results.alerts.length > 0}
              <Command.Group heading="Alerts">
                {#each results.alerts as item (item.route)}
                  <Command.Item
                    value={item.route}
                    onSelect={() => handleSelect(item)}
                  >
                    <CommandPaletteItem {item} />
                  </Command.Item>
                {/each}
              </Command.Group>
            {/if}
          {/if}
        </Command.List>

        <div class="command-palette-footer">
          <span class="command-palette-footer-hint">
            <kbd>↑</kbd><kbd>↓</kbd> navigate
          </span>
          <span class="command-palette-footer-hint">
            <kbd>↵</kbd> open
          </span>
          <span class="command-palette-footer-action">
            <kbd>{isMac ? "⌘" : "Ctrl+"}K</kbd>
            open / close
          </span>
        </div>
      </Command.Root>
    </div>
  </Dialog.Content>
</Dialog.Root>

<style>
  /* Override the dialog content to remove default styling */
  :global(.command-palette-dialog) {
    padding: 0 !important;
    border: none !important;
    background: transparent !important;
    box-shadow: none !important;
    max-width: 640px !important;
    width: 100% !important;
  }

  .command-palette {
    width: 100%;
  }

  /* Root container: uses Rill semantic tokens */
  .command-palette :global([data-cmdk-root]) {
    max-width: 640px;
    width: 100%;
    background: var(--surface-overlay);
    border-radius: 12px;
    padding: 8px 0;
    font-family: var(--font-sans, system-ui, -apple-system, sans-serif);
    border: 1px solid var(--border);
    box-shadow:
      0 16px 70px rgba(0, 0, 0, 0.15),
      0 0 0 1px rgba(0, 0, 0, 0.05);
    position: relative;
  }

  /* Input */
  .command-palette :global([data-cmdk-input]) {
    border: none;
    width: 100%;
    font-size: 15px;
    padding: 8px 16px;
    outline: none;
    background: transparent;
    color: var(--fg-primary);
    font-family: inherit;
  }

  .command-palette :global([data-cmdk-input])::placeholder {
    color: var(--fg-muted);
  }

  /* Separator between input and list */
  .command-palette-separator {
    height: 1px;
    background: var(--border);
    margin: 8px 0;
  }

  /* List */
  .command-palette :global([data-cmdk-list]) {
    padding: 0 8px;
    max-height: 340px;
    overflow: auto;
    overscroll-behavior: contain;
    transition: height 100ms ease;
    padding-bottom: 40px;
  }

  /* Items */
  .command-palette :global([data-cmdk-item]) {
    cursor: pointer;
    height: 40px;
    border-radius: 8px;
    font-size: 14px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 8px;
    color: var(--fg-primary);
    user-select: none;
    will-change: background, color;
    transition: all 150ms ease;
    transition-property: none;
  }

  .command-palette :global([data-cmdk-item][data-selected="true"]) {
    background: var(--surface-hover);
  }

  .command-palette :global([data-cmdk-item]:active) {
    transition-property: background;
    background: var(--surface-active);
  }

  .command-palette :global([data-cmdk-item]:first-child) {
    margin-top: 8px;
  }

  .command-palette :global([data-cmdk-item] + [data-cmdk-item]) {
    margin-top: 4px;
  }

  /* Group headings */
  .command-palette :global([data-cmdk-group-heading]) {
    user-select: none;
    font-size: 12px;
    color: var(--fg-muted);
    padding: 0 8px;
    display: flex;
    align-items: center;
    margin-bottom: 4px;
  }

  .command-palette :global(*:not([hidden]) + [data-cmdk-group]) {
    margin-top: 8px;
  }

  /* Empty state */
  .command-palette :global([data-cmdk-empty]),
  .command-palette-status {
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 64px;
    white-space: pre-wrap;
    color: var(--fg-muted);
  }

  /* Footer */
  .command-palette-footer {
    display: flex;
    height: 40px;
    align-items: center;
    width: 100%;
    padding: 8px 16px;
    border-top: 1px solid var(--border);
    gap: 12px;
    font-size: 12px;
    color: var(--fg-tertiary);
  }

  .command-palette-footer-hint {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .command-palette-footer-action {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-left: auto;
  }

  .command-palette-footer :global(kbd) {
    background: var(--surface-muted);
    color: var(--fg-secondary);
    height: 20px;
    min-width: 20px;
    border-radius: 4px;
    padding: 0 4px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-family: inherit;
    font-size: 11px;
  }
</style>
