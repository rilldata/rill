<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { browser } from "$app/environment";
  import {
    Dialog as CommandDialog,
    Input as CommandInput,
    List as CommandList,
    Empty as CommandEmpty,
    Group as CommandGroup,
    Item as CommandItem,
  } from "@rilldata/web-common/components/command";
  import { createAdminServiceListProjectsForOrganizationAndUser } from "@rilldata/web-admin/client";
  import { searchIndex } from "./search-orchestrator";
  import { buildRoute } from "./route-builders";
  import CommandPaletteItem from "./CommandPaletteItem.svelte";
  import type { SearchableItem } from "./types";

  let open = false;
  let query = "";

  $: orgName = $page.params.organization;
  $: isMac = browser && window.navigator.userAgent.includes("Macintosh");

  // Prefetch project list
  $: projectListQuery = createAdminServiceListProjectsForOrganizationAndUser(
    orgName,
    { pageSize: 50 },
    {
      query: {
        enabled: !!orgName,
        staleTime: 5 * 60 * 1000,
      },
    },
  );

  // Build search index from projects
  $: projectItems = buildProjectItems(orgName, $projectListQuery.data?.projects);

  // Resource items will be populated by Task 6 (resource prefetch)
  let resourceItems: SearchableItem[] = [];

  $: searchItems = [...projectItems, ...resourceItems];
  $: results = searchIndex(searchItems, query);
  $: hasResults =
    results.projects.length > 0 ||
    results.dashboards.length > 0 ||
    results.reports.length > 0 ||
    results.alerts.length > 0;

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

<CommandDialog bind:open shouldFilter={false}>
  <CommandInput
    placeholder="Search projects, dashboards, reports..."
    bind:value={query}
  />
  <CommandList>
    {#if query.length < 2}
      <div class="py-6 text-center text-sm text-gray-500">
        Type to search...
      </div>
    {:else if !hasResults}
      <CommandEmpty>No results found.</CommandEmpty>
    {:else}
      {#if results.projects.length > 0}
        <CommandGroup heading="Projects">
          {#each results.projects as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}

      {#if results.dashboards.length > 0}
        <CommandGroup heading="Dashboards">
          {#each results.dashboards as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}

      {#if results.reports.length > 0}
        <CommandGroup heading="Reports">
          {#each results.reports as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}

      {#if results.alerts.length > 0}
        <CommandGroup heading="Alerts">
          {#each results.alerts as item (item.route)}
            <CommandItem
              value={item.route}
              onSelect={() => handleSelect(item)}
            >
              <CommandPaletteItem {item} />
            </CommandItem>
          {/each}
        </CommandGroup>
      {/if}
    {/if}
  </CommandList>

  <div
    class="flex items-center gap-4 px-4 py-2 border-t border-gray-700 text-[11px] text-gray-500"
  >
    <span>↑↓ navigate</span>
    <span>↵ open</span>
    <span class="ml-auto flex items-center gap-1">
      <kbd
        class="bg-gray-800 border border-gray-600 rounded px-1.5 py-0.5 text-[10px] text-gray-400"
      >
        {isMac ? "⌘" : "Ctrl+"}K
      </kbd>
      open / close menu
    </span>
  </div>
</CommandDialog>
