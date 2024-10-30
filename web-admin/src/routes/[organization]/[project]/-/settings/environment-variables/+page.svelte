<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProjectVariables } from "@rilldata/web-admin/client";
  import EnvironmentVariablesTable from "@rilldata/web-admin/features/projects/environment-variables/EnvironmentVariablesTable.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";

  let open = false;
  let searchText = "";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: getProjectVariables = createAdminServiceGetProjectVariables(
    organization,
    project,
  );

  $: projectVariables = $getProjectVariables.data?.variables || [];

  // TODO: wire up searchText to filter variables key
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
              href="https://docs.rilldata.com/tutorials/administration/project/credential-envvariable-mangement"
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
            disabled={searchText === ""}
          />
          <!-- TODO: filter by environment -->
          <Button type="primary" large on:click={() => (open = true)}>
            <Plus size="16px" />
            <span>Add environment variable</span>
          </Button>
        </div>
        <EnvironmentVariablesTable data={projectVariables} />
      </div>
    {/if}
  </div>
</div>
