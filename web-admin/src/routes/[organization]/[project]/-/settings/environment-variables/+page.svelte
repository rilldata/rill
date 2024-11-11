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
  import type { EnvironmentTypes } from "@rilldata/web-admin/features/projects/environment-variables/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceAnalyzeVariables } from "@rilldata/web-common/runtime-client";

  let open = false;
  let searchText = "";
  let isDropdownOpen = false;
  let filterByEnvironment: EnvironmentTypes = "";

  // TODO: revisit this
  // AnalyzeVariables scans Source, Model and Connector resources in the catalog for use of an environment variable
  // TODO: how do i use the variables in the cloud?
  $: ({ instanceId } = $runtime);
  $: analyzeVariables = createRuntimeServiceAnalyzeVariables(instanceId);

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

  $: variableNames = projectVariables.map((variable) => variable.name);

  $: searchedVariables = projectVariables.filter((variable) =>
    variable.name.toLowerCase().includes(searchText.toLowerCase()),
  );

  $: filteredVariables = searchedVariables.filter((variable) => {
    if (filterByEnvironment === "") return true;
    return variable.environment === filterByEnvironment;
  });

  function handleFilterByEnvironment(environment: EnvironmentTypes) {
    filterByEnvironment = environment;
  }

  $: environmentLabel =
    filterByEnvironment === ""
      ? "All environments"
      : filterByEnvironment === "prod"
        ? "Production"
        : "Development";

  $: emptyTextWhenNoVariables =
    filterByEnvironment === ""
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
          <p class="text-base font-normal text-slate-700">
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
              class="min-w-fit flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
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
                checked={filterByEnvironment === ""}
                on:click={() => handleFilterByEnvironment("")}
              >
                All environments
              </DropdownMenu.CheckboxItem>
              <DropdownMenu.CheckboxItem
                checked={filterByEnvironment === "prod"}
                on:click={() => handleFilterByEnvironment("prod")}
              >
                Production
              </DropdownMenu.CheckboxItem>
              <DropdownMenu.CheckboxItem
                checked={filterByEnvironment === "dev"}
                on:click={() => handleFilterByEnvironment("dev")}
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
          data={filteredVariables}
          emptyText={emptyTextWhenNoVariables}
        />
      </div>
    {/if}
  </div>
</div>

<AddDialog bind:open {variableNames} />
