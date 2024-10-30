<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProjectVariables } from "@rilldata/web-admin/client";
  import Empty from "@rilldata/web-admin/features/projects/environment-variables/Empty.svelte";
  import EnvironmentVariablesTable from "@rilldata/web-admin/features/projects/environment-variables/EnvironmentVariablesTable.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: getProjectVariables = createAdminServiceGetProjectVariables(
    organization,
    project,
  );

  $: projectVariablesMap = $getProjectVariables.data || [];
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
      {#if $getProjectVariables.data.variables.length === 0}
        <Empty />
      {:else}
        <!-- TODO: use variablesMap -->
        <EnvironmentVariablesTable data={$getProjectVariables.data.variables} />
      {/if}
    {/if}
  </div>
</div>
