<script lang="ts">
  import {
    createAdminServiceListProjectsForOrganization,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
    ViewMode,
  } from "@rilldata/web-common/components/table-toolbar";
  import { projectWelcomeEnabled } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
  import { Plus } from "lucide-svelte";
  import ProjectCards from "./ProjectCards.svelte";
  import ProjectsTable from "./ProjectsTable.svelte";

  export let organization: string;

  let viewMode: ViewMode = "grid";
  let searchText = "";
  let sortDirection: SortDirection = "newest";
  let permissionSelected: string[] = [];

  $: projs = createAdminServiceListProjectsForOrganization(organization, {
    pageSize: 1000,
  });
  $: projects = $projs.data?.projects ?? [];

  $: filterGroups = [
    {
      label: "Permission",
      key: "permission",
      options: [
        { value: "public", label: "Public" },
        { value: "private", label: "Private" },
      ],
      selected: permissionSelected,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[];

  function onFilterChange(key: string, selected: string | string[]) {
    if (key === "permission") permissionSelected = selected as string[];
  }

  function onClearAllFilters() {
    permissionSelected = [];
    searchText = "";
  }

  function matchesSearch(p: V1Project, q: string): boolean {
    if (!q) return true;
    const needle = q.toLowerCase();
    return (
      (p.name ?? "").toLowerCase().includes(needle) ||
      (p.description ?? "").toLowerCase().includes(needle)
    );
  }

  function matchesPermission(p: V1Project, selected: string[]): boolean {
    if (selected.length === 0) return true;
    const isPublic = !!p.public;
    return (
      (isPublic && selected.includes("public")) ||
      (!isPublic && selected.includes("private"))
    );
  }

  function compareByCreated(a: V1Project, b: V1Project, dir: SortDirection) {
    const aTime = a.createdOn ? new Date(a.createdOn).getTime() : 0;
    const bTime = b.createdOn ? new Date(b.createdOn).getTime() : 0;
    return dir === "newest" ? bTime - aTime : aTime - bTime;
  }

  $: visibleProjects = projects
    .filter(
      (p) =>
        matchesSearch(p, searchText) &&
        matchesPermission(p, permissionSelected),
    )
    .sort((a, b) => compareByCreated(a, b, sortDirection));
</script>

<div class="flex flex-col gap-y-2 w-full">
  <div class="flex items-center h-16">
    <h2 class="grow text-2xl font-semibold text-fg-secondary leading-9">
      Projects
    </h2>
    {#if projectWelcomeEnabled}
      <Button type="secondary" href="/{organization}/-/create-project">
        <Plus size="16" />
        New project
      </Button>
    {/if}
  </div>

  <TableToolbar
    bind:searchText
    bind:sortDirection
    bind:viewMode
    showViewToggle
    {filterGroups}
    {onFilterChange}
    {onClearAllFilters}
  />

  {#if projects.length === 0}
    <p class="text-fg-secondary text-xs">
      This organization has no projects yet.
    </p>
  {:else if visibleProjects.length === 0}
    <p class="text-fg-secondary text-sm py-4">
      No projects match your filters.
    </p>
  {:else if viewMode === "grid"}
    <ProjectCards {organization} projects={visibleProjects} />
  {:else}
    <ProjectsTable {organization} projects={visibleProjects} />
  {/if}
</div>
