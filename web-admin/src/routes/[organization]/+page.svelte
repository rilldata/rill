<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { createAdminServiceGetOrganization } from "../../client";
  import OrganizationHero from "../../features/organizations/OrganizationHero.svelte";
  import ProjectCards from "../../features/projects/ProjectCards.svelte";

  export let data;
  $: ({ organizationPermissions } = data);
  $: ({
    params: { organization: orgName },
  } = $page);

  $: org = createAdminServiceGetOrganization(orgName);

  $: title = $org.data?.organization?.displayName || orgName;
</script>

<svelte:head>
  <title>{m.organizations_overview_page_title({ title })}</title>
</svelte:head>

<ContentContainer showTitle={false} maxWidth={1300}>
  <OrganizationHero {title} />
  <ProjectCards
    organization={orgName}
    createProjectsPermission={organizationPermissions.createProjects}
  />
</ContentContainer>
