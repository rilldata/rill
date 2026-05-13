<script lang="ts">
  import { page } from "$app/state";
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
  import { Plus } from "lucide-svelte";
  import Callout from "web-common/src/components/callout/Callout.svelte";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";

  let {
    searchText = $bindable(""),
    envFilter = $bindable<EnvironmentTypes[]>([]),
  }: {
    searchText?: string;
    envFilter?: EnvironmentTypes[];
  } = $props();

  let open = $state(false);

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);
  let activeBranch = $derived(extractBranchFromPath(page.url.pathname));

  let getProjectVariables = $derived(
    createAdminServiceGetProjectVariables(organization, project, {
      forAllEnvironments: true,
    }),
  );

  let projectVariables = $derived($getProjectVariables.data?.variables || []);

  let variableNames = $derived(
    projectVariables.map((variable) => {
      return {
        environment: getEnvironmentType(variable.environment),
        name: variable.name,
      };
    }),
  );

  let searchedVariables = $derived(
    projectVariables.filter((variable) =>
      variable.name.toLowerCase().includes(searchText.toLowerCase()),
    ),
  );

  let filteredVariables = $derived(
    searchedVariables.filter((variable) => {
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
    }),
  );

  let sortedVariables = $derived(
    [...filteredVariables].sort((a, b) => {
      return new Date(b.updatedOn).getTime() - new Date(a.updatedOn).getTime();
    }),
  );

  function handleFilterChange(_key: string, selected: string | string[]) {
    envFilter = selected as EnvironmentTypes[];
  }

  function handleClearAllFilters() {
    envFilter = [];
    searchText = "";
  }

  let emptyTextWhenNoVariables = $derived(
    envFilter.length === 0
      ? "No environment variables"
      : `No environment variables match the selected filters`,
  );

  let filterGroups = $derived([
    {
      label: "Environment",
      key: "environment",
      options: [
        { value: EnvironmentType.PRODUCTION, label: "Production" },
        { value: EnvironmentType.DEVELOPMENT, label: "Development" },
      ],
      selected: envFilter,
      defaultValue: [],
      multiSelect: true,
    },
  ]);
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
      <div class="flex flex-col gap-3 w-full overflow-hidden">
        <div class="flex flex-col">
          <RadixLarge>Environment variables</RadixLarge>

          <p class="text-sm text-fg-tertiary font-medium">
            Manage your environment variables here. Note that these are project
            wide. <a
              href="https://docs.rilldata.com/guide/administration/project-settings/variables-and-credentials"
              target="_blank"
              class="text-primary-600 hover:text-primary-700 active:text-primary-800"
            >
              Learn more ->
            </a>
          </p>

          {#if activeBranch}
            <Callout level="info">
              <span class="text-sm">
                These settings apply to the entire project, not just this
                branch.
              </span>
            </Callout>
          {/if}
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
            <Plus size="16px" /> New key
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
