<script lang="ts">
  import { page } from "$app/state";
  import { createAdminServiceGetProjectVariables } from "@rilldata/web-admin/client";
  import AddDialog from "@rilldata/web-admin/features/projects/environment-variables/AddDialog.svelte";
  import EnvironmentVariablesTable from "@rilldata/web-admin/features/projects/environment-variables/EnvironmentVariablesTable.svelte";
  import { EnvironmentType } from "@rilldata/web-admin/features/projects/environment-variables/types";
  import { getEnvironmentType } from "@rilldata/web-admin/features/projects/environment-variables/utils";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import {
    type FilterGroup,
    TableToolbar,
  } from "@rilldata/web-common/components/table-toolbar";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Plus } from "lucide-svelte";
  import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
  import { DebouncedRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  let open = $state(false);

  const searchTextStore = new DebouncedRuneStore(
    UrlParamsState.createStringParam("q"),
    500,
  );
  const envFilterStore = UrlParamsState.createStringArrayParam("env");

  let { organization, project } = $derived(page.params);

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
      variable.name.toLowerCase().includes(searchTextStore.value.toLowerCase()),
    ),
  );

  let filteredVariables = $derived(
    searchedVariables.filter((variable) => {
      if (envFilterStore.value.length === 0) return true;
      return envFilterStore.value.some((sel) => {
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

  function handleClearAllFilters() {
    searchTextStore.immediateSetter("");
    envFilterStore.setter([]);
  }

  let emptyTextWhenNoVariables = $derived(
    envFilterStore.value.length === 0
      ? m.env_no_variables()
      : m.env_no_match_filters(),
  );

  let filterGroups = $derived<FilterGroup[]>([
    {
      label: m.env_environment_label(),
      key: "environment",
      options: [
        { value: EnvironmentType.PRODUCTION, label: m.env_production_label() },
        {
          value: EnvironmentType.DEVELOPMENT,
          label: m.env_development_label(),
        },
      ],
      selectedStore: envFilterStore,
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
        {m.env_error_loading()}
        {$getProjectVariables.error}
      </div>
    {:else if $getProjectVariables.isSuccess}
      <div class="flex flex-col gap-3 w-full overflow-hidden">
        <div class="flex flex-col">
          <RadixLarge>{m.env_page_title()}</RadixLarge>
          <p class="text-sm text-fg-tertiary font-medium">
            {m.env_page_description()}
            <a
              href="https://docs.rilldata.com/guide/administration/project-settings/variables-and-credentials"
              target="_blank"
              class="text-primary-600 hover:text-primary-700 active:text-primary-800"
            >
              {m.env_learn_more()}
            </a>
          </p>
        </div>
        <TableToolbar
          {searchTextStore}
          {filterGroups}
          onClearAllFilters={handleClearAllFilters}
        >
          <Button type="primary" large onClick={() => (open = true)}>
            <Plus size="16px" />
            {m.env_new_key_button()}
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
