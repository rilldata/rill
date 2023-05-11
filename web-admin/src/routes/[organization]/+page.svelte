<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    createAdminServiceGetOrganization,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";

  $: org = createAdminServiceGetOrganization($page.params.organization);
  $: projs = createAdminServiceListProjectsForOrganization(
    $page.params.organization,
    undefined,
    { query: { enabled: !!$org.data.organization } }
  );

  $: if ($projs.data && $projs.data.projects?.length > 0) {
    goto(
      `/${$page.params.organization}/${$projs.data.projects[0].name}/-/redirect`
    );
  }

  function openDocs() {
    window.open("https://docs.rilldata.com/quickstart/local", "_blank");
  }
</script>

<svelte:head>
  <title>{$page.params.organization} overview - Rill</title>
</svelte:head>

<div class="flex-grow overflow-hidden">
  <section class="flex flex-col justify-center items-center h-3/5">
    {#if $org.isLoading || $projs.isLoading}
      <span>Loading...</span>
    {:else if $org.isError || $projs.isError}
      <span>Error: {$org.error || $projs.error}</span>
    {:else if $org.data && $org.data.organization && $projs.data && $projs.data.projects?.length === 0}
      <h1 class="text-3xl font-medium text-gray-800 mb-4">
        Organization: {$org.data.organization.name}
      </h1>
      <p class="text-lg text-gray-700 mb-6">
        Your organization does not have any projects... yet!
      </p>
      <Button type="primary" on:click={openDocs}>Read the docs</Button>
    {/if}
  </section>
</div>
