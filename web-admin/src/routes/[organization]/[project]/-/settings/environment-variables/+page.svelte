<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProjectVariables } from "@rilldata/web-admin/client";
  import AddDialog from "@rilldata/web-admin/features/projects/environment-variables/AddDialog.svelte";
  import EnvironmentVariablesTable from "@rilldata/web-admin/features/projects/environment-variables/EnvironmentVariablesTable.svelte";
  import {
    EnvironmentType,
    type EnvironmentTypes,
  } from "@rilldata/web-admin/features/projects/environment-variables/types";
  import { getEnvironmentType } from "@rilldata/web-admin/features/projects/environment-variables/utils";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type { SortDirection } from "@rilldata/web-common/components/table-toolbar/types";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { Plus } from "lucide-svelte";

  let open = false;
  let searchText = "";
  let filterByEnvironment: EnvironmentTypes = EnvironmentType.UNDEFINED;
  let sortDirection: SortDirection = "newest";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: getProjectVariables = createAdminServiceGetProjectVariables(
    organization,
    project,
    {
      forAllEnvironments: true,
    },
  );

  $: projectVariables = $getProjectVariables.data?.variables || [];

  $: variableNames = projectVariables.map((variable) => {
    return {
      environment: getEnvironmentType(variable.environment),
      name: variable.name,
    };
  });

  $: searchedVariables = projectVariables.filter((variable) =>
    variable.name.toLowerCase().includes(searchText.toLowerCase()),
  );

  $: filteredVariables = searchedVariables.filter((variable) => {
    if (filterByEnvironment === EnvironmentType.UNDEFINED) {
      return true;
    }
    if (filterByEnvironment === EnvironmentType.DEVELOPMENT) {
      return (
        variable.environment === EnvironmentType.DEVELOPMENT ||
        variable.environment === EnvironmentType.UNDEFINED
      );
    }
    if (filterByEnvironment === EnvironmentType.PRODUCTION) {
      return (
        variable.environment === EnvironmentType.PRODUCTION ||
        variable.environment === EnvironmentType.UNDEFINED
      );
    }
    return false;
  });

  $: sortedVariables = [...filteredVariables].sort((a, b) => {
    const aTime = new Date(a.updatedOn).getTime();
    const bTime = new Date(b.updatedOn).getTime();
    return sortDirection === "newest" ? bTime - aTime : aTime - bTime;
  });

  function handleFilterChange(_key: string, value: string) {
    filterByEnvironment = value as EnvironmentTypes;
  }

  function handleSortToggle() {
    sortDirection = sortDirection === "newest" ? "oldest" : "newest";
  }

  function handleClearAllFilters() {
    filterByEnvironment = EnvironmentType.UNDEFINED;
  }

  $: environmentLabel =
    filterByEnvironment === EnvironmentType.UNDEFINED
      ? "All environments"
      : filterByEnvironment === EnvironmentType.PRODUCTION
        ? "Production"
        : "Development";

  $: emptyTextWhenNoVariables =
    filterByEnvironment === EnvironmentType.UNDEFINED
      ? "No environment variables"
      : `No environment variables for ${environmentLabel}`;

  $: filterGroups = [
    {
      label: "Filter by environment",
      key: "environment",
      options: [
        { value: EnvironmentType.UNDEFINED, label: "All environments" },
        { value: EnvironmentType.PRODUCTION, label: "Production" },
        { value: EnvironmentType.DEVELOPMENT, label: "Development" },
      ],
      selected: filterByEnvironment,
      defaultValue: EnvironmentType.UNDEFINED,
    },
  ];
</script>

<div class="flex flex-col w-full overflow-hidden">
  <div class="flex md:flex-row flex-col gap-6">
    {#if $getProjectVariables.isLoading}
      <DelayedSpinner isLoading={$getProjectVariables.isLoading} size="1rem" />
    {:else if $getProjectVariables.isError}
      <div class="text-red-500">
        Error loading environment variables: {$getProjectVariables.error}
      </div>
    {:else if $getProjectVariables.isSuccess}
      <div class="flex flex-col gap-6 w-full overflow-hidden">
        <div class="flex flex-col">
          <RadixLarge>Environment variables</RadixLarge>
          <p class="text-sm text-fg-tertiary font-medium">
            Manage your environment variables here. <a
              href="https://docs.rilldata.com/guide/administration/project-settings/variables-and-credentials"
              target="_blank"
              class="text-primary-600 hover:text-primary-700 active:text-primary-800"
            >
              Learn more ->
            </a>
          </p>
        </div>
        <TableToolbar
          {searchText}
          onSearchChange={(text) => (searchText = text)}
          searchDisabled={projectVariables.length === 0}
          {filterGroups}
          onFilterChange={handleFilterChange}
          onClearAllFilters={handleClearAllFilters}
          {sortDirection}
          onSortToggle={handleSortToggle}
        >
          <Button type="primary" large onClick={() => (open = true)}>
            <Plus size="16px" />
          </Button>
        </TableToolbar>
        <EnvironmentVariablesTable
          data={sortedVariables}
          emptyText={emptyTextWhenNoVariables}
          {variableNames}
        />
      </div>
    {/if}
  </div>
</div>

<AddDialog bind:open {variableNames} />
