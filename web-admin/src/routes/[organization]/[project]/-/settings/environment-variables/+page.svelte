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
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Plus } from "lucide-svelte";
  import { onMount } from "svelte";

  let open = false;

  // Filters — synced to URL params `q` and `env` (multi-select array)
  const filterSync = createUrlFilterSync([
    { key: "q", type: "string" },
    { key: "env", type: "array" },
  ]);
  filterSync.init($page.url);

  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let envFilter: EnvironmentTypes[] = parseArrayParam(
    $page.url.searchParams.get("env"),
  ) as EnvironmentTypes[];
  let mounted = false;

  // URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    searchText = parseStringParam($page.url.searchParams.get("q"));
    envFilter = parseArrayParam(
      $page.url.searchParams.get("env"),
    ) as EnvironmentTypes[];
  }

  // Local state → URL
  $: if (mounted) {
    filterSync.syncToUrl({ q: searchText, env: envFilter });
  }

  onMount(() => {
    mounted = true;
  });

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
    if (envFilter.length === 0) return true;
    return envFilter.some((sel) => {
      if (sel === EnvironmentType.DEVELOPMENT) {
        return (
          variable.environment === EnvironmentType.DEVELOPMENT ||
          variable.environment === EnvironmentType.UNDEFINED
        );
      }
      if (sel === EnvironmentType.PRODUCTION) {
        return (
          variable.environment === EnvironmentType.PRODUCTION ||
          variable.environment === EnvironmentType.UNDEFINED
        );
      }
      return false;
    });
  });

  $: sortedVariables = [...filteredVariables].sort((a, b) => {
    return new Date(b.updatedOn).getTime() - new Date(a.updatedOn).getTime();
  });

  function handleFilterChange(_key: string, selected: string | string[]) {
    envFilter = selected as EnvironmentTypes[];
  }

  function handleClearAllFilters() {
    envFilter = [];
    searchText = "";
  }

  $: emptyTextWhenNoVariables =
    envFilter.length === 0
      ? m.env_no_variables()
      : m.env_no_match_filters();

  $: filterGroups = [
    {
      label: m.env_environment_label(),
      key: "environment",
      options: [
        { value: EnvironmentType.PRODUCTION, label: m.env_production_label() },
        { value: EnvironmentType.DEVELOPMENT, label: m.env_development_label() },
      ],
      selected: envFilter,
      defaultValue: [],
      multiSelect: true,
    },
  ];
</script>

<div class="flex flex-col w-full overflow-hidden">
  <div class="flex md:flex-row flex-col gap-6">
    {#if $getProjectVariables.isLoading}
      <DelayedSpinner isLoading={$getProjectVariables.isLoading} size="1rem" />
    {:else if $getProjectVariables.isError}
      <div class="text-red-500">
        {m.env_error_loading()} {$getProjectVariables.error}
      </div>
    {:else if $getProjectVariables.isSuccess}
      <div class="flex flex-col gap-3 w-full overflow-hidden">
        <div class="flex flex-col">
          <RadixLarge>{m.env_page_title()}</RadixLarge>
          <p class="text-sm text-fg-tertiary font-medium">
            {m.env_page_description()} <a
              href="https://docs.rilldata.com/guide/administration/project-settings/variables-and-credentials"
              target="_blank"
              class="text-primary-600 hover:text-primary-700 active:text-primary-800"
            >
              {m.env_learn_more()}
            </a>
          </p>
        </div>
        <TableToolbar
          bind:searchText
          searchDisabled={projectVariables.length === 0}
          {filterGroups}
          onFilterChange={handleFilterChange}
          onClearAllFilters={handleClearAllFilters}
          showSort={false}
        >
          <Button type="primary" large onClick={() => (open = true)}>
            <Plus size="16px" /> {m.env_new_key_button()}
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
