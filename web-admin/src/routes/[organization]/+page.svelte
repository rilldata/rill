<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
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

<ContentContainer showTitle={false} maxWidth={1300}>
  {#if $org.data && $org.data.organization && $projs.data}
    {#if $projs.data.projects?.length === 0}
      <OrganizationHero {title} />
      <span>
        This organization has no projects yet. <a
          href="https://docs.rilldata.com/"
          target="_blank"
          rel="noreferrer noopener">See docs</a
        >
      </span>
    {:else if $allProjectsHibernating.data}
      <OrganizationHibernating
        organization={orgName}
        {organizationPermissions}
      />
    {:else}
      <div class="flex flex-col gap-y-8">
        <OrganizationHero {title} />
        <ProjectCards organization={orgName} />
      </div>
    {/if}
  {/if}
</ContentContainer>
