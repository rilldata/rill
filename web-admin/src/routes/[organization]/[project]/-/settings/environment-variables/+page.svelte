<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProjectVariables } from "@rilldata/web-admin/client";
  import EnvironmentVariablesTable from "@rilldata/web-admin/features/projects/environment-variables/EnvironmentVariablesTable.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import AddDialog from "@rilldata/web-admin/features/projects/environment-variables/AddDialog.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    EnvironmentType,
    type EnvironmentTypes,
  } from "@rilldata/web-admin/features/projects/environment-variables/types";
  import { getEnvironmentType } from "@rilldata/web-admin/features/projects/environment-variables/utils";

  let open = false;
  let searchText = "";
  let isDropdownOpen = false;
  let filterByEnvironment: EnvironmentTypes = EnvironmentType.UNDEFINED;

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
    // Show all variables
    if (filterByEnvironment === EnvironmentType.UNDEFINED) {
      return true;
    }
    // Includes development
    if (filterByEnvironment === EnvironmentType.DEVELOPMENT) {
      return (
        variable.environment === EnvironmentType.DEVELOPMENT ||
        variable.environment === EnvironmentType.UNDEFINED
      );
    }
    // Includes production
    if (filterByEnvironment === EnvironmentType.PRODUCTION) {
      return (
        variable.environment === EnvironmentType.PRODUCTION ||
        variable.environment === EnvironmentType.UNDEFINED
      );
    }
    // No match
    return false;
  });

  $: sortedVariables = filteredVariables.sort((a, b) => {
    return new Date(b.updatedOn).getTime() - new Date(a.updatedOn).getTime();
  });

  function handleFilterByEnvironment(environment: EnvironmentTypes) {
    filterByEnvironment = environment;
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
</script>

<div class="flex flex-col w-full">
  <div class="flex md:flex-row flex-col gap-6">
    {#if $getProjectVariables.isLoading}
      <DelayedSpinner isLoading={$getProjectVariables.isLoading} size="1rem" />
    {:else if $getProjectVariables.isError}
      <div class="text-red-500">
        Error loading environment variables: {$getProjectVariables.error}
      </div>
    {:else if $getProjectVariables.isSuccess}
      <div class="flex flex-col gap-6 w-full">
        <div class="flex flex-col">
          <RadixLarge>Environment variables</RadixLarge>
          <p class="text-sm text-slate-700 font-medium">
            Manage your environment variables here. <a
              href="https://docs.rilldata.com/tutorials/administration/project/credentials-env-variable-management"
              target="_blank"
              class="text-primary-600 hover:text-primary-700 active:text-primary-800"
            >
              Learn more ->
            </a>
          </p>
        </div>
        <div class="flex flex-row gap-x-4">
          <Search
            placeholder="Search"
            bind:value={searchText}
            large
            autofocus={false}
            showBorderOnFocus={false}
            disabled={projectVariables.length === 0}
          />
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              class="min-w-fit flex flex-row gap-1 items-center rounded-sm border border-slate-300 {isDropdownOpen
                ? 'bg-slate-200'
                : 'hover:bg-slate-100'} px-2 py-1"
            >
              <span class="text-slate-600 font-medium">{environmentLabel}</span>
              {#if isDropdownOpen}
                <CaretUpIcon size="12px" />
              {:else}
                <CaretDownIcon size="12px" />
              {/if}
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <DropdownMenu.Label class="uppercase"
                >Filter by environment</DropdownMenu.Label
              >
              <DropdownMenu.CheckboxItem
                checked={filterByEnvironment === EnvironmentType.UNDEFINED}
                on:click={() =>
                  handleFilterByEnvironment(EnvironmentType.UNDEFINED)}
              >
                All environments
              </DropdownMenu.CheckboxItem>
              <DropdownMenu.CheckboxItem
                checked={filterByEnvironment === EnvironmentType.PRODUCTION}
                on:click={() =>
                  handleFilterByEnvironment(EnvironmentType.PRODUCTION)}
              >
                Production
              </DropdownMenu.CheckboxItem>
              <DropdownMenu.CheckboxItem
                checked={filterByEnvironment === EnvironmentType.DEVELOPMENT}
                on:click={() =>
                  handleFilterByEnvironment(EnvironmentType.DEVELOPMENT)}
              >
                Development
              </DropdownMenu.CheckboxItem>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
          <Button type="primary" large on:click={() => (open = true)}>
            <Plus size="16px" />
            <span>Add environment variable</span>
          </Button>
        </div>
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
