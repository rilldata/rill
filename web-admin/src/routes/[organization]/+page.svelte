<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { createAdminServiceGetOrganization } from "../../client";
  import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options";
  import { createQuery } from "@tanstack/svelte-query";
  import OrganizationHero from "../../features/organizations/OrganizationHero.svelte";
  import ProjectCards from "../../features/projects/ProjectCards.svelte";

  export let data;
  $: ({ organizationPermissions } = data);
  $: ({
    params: { organization: orgName },
  } = $page);

  $: org = createAdminServiceGetOrganization(orgName);
  $: projs = createQuery({
    ...listProjectsForOrgQueryOptions(orgName),
    enabled: !!$org.data?.organization,
  });

  $: title = $org.data?.organization?.displayName || orgName;
</script>

<svelte:head>
  <title>{title} overview - Rill</title>
</svelte:head>

<ContentContainer showTitle={false} maxWidth={1300}>
  <OrganizationHero {title} />
  <ProjectCards
    organization={orgName}
    createProjectsPermission={organizationPermissions.createProjects}
  />
</ContentContainer>
