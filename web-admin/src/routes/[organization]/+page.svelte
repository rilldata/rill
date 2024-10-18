<script lang="ts">
  import { page } from "$app/stores";
  import OrganizationHibernating from "@rilldata/web-admin/features/organizations/hibernating/OrganizationHibernating.svelte";
  import { areAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
  import {
    createAdminServiceGetOrganization,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import OrganizationHero from "../../features/organizations/OrganizationHero.svelte";
  import ProjectCards from "../../features/projects/ProjectCards.svelte";

  export let data;
  $: ({ organizationPermissions } = data);
  $: ({
    params: { organization: orgName },
  } = $page);

  $: org = createAdminServiceGetOrganization(orgName);
  $: projs = createAdminServiceListProjectsForOrganization(orgName, undefined, {
    query: { enabled: !!$org.data?.organization },
  });
  $: allProjectsHibernating = areAllProjectsHibernating(orgName);

  $: title = $org.data?.organization?.displayName || orgName;
</script>

<svelte:head>
  <title>{title} overview - Rill</title>
</svelte:head>

{#if $org.data && $org.data.organization && $projs.data}
  <section
    class="mx-8 my-8 sm:my-16 sm:mx-16 lg:mx-32 lg:my-24 2xl:mx-64 mx-auto flex flex-col gap-y-4"
  >
    {#if $allProjectsHibernating.data}
      <OrganizationHibernating
        organization={orgName}
        {organizationPermissions}
      />
    {:else}
      <OrganizationHero organization={orgName} {title} />
      {#if $projs.data.projects?.length === 0}
        <span
          >This organization has no projects yet. <a
            href="https://docs.rilldata.com/home/get-started"
            target="_blank"
            rel="noreferrer noopener">See docs</a
          ></span
        >
      {:else}
        <div class="py-2 px-1.5">
          <ProjectCards organization={orgName} />
        </div>
      {/if}
    {/if}
  </section>
{/if}
