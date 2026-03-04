<script lang="ts">
  import { page } from "$app/stores";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizations as listOrgs,
    createAdminServiceListProjectsForOrganization as listProjects,
    type V1Organization,
  } from "../../client";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import SignIn from "../authentication/SignIn.svelte";
  import { isOrganizationPage } from "../navigation/nav-utils";

  export let readProjects: boolean;
  export let planDisplayName: string | undefined;
  export let organizationLogoUrl: string | undefined;

  const user = createAdminServiceGetCurrentUser();

  $: ({
    params: { organization, project },
  } = $page);

  $: onOrgPage = isOrganizationPage($page);

  $: loggedIn = !!$user.data?.user;
  $: rillLogoHref = !loggedIn ? "https://www.rilldata.com" : "/";

  $: organizationQuery = listOrgs(
    { pageSize: 100 },
    {
      query: {
        enabled: !!$user.data?.user,
        retry: 2,
        refetchOnMount: true,
      },
    },
  );

  $: projectsQuery = listProjects(
    organization,
    { pageSize: 100 },
    {
      query: {
        enabled: !!organization && readProjects,
        retry: 2,
        refetchOnMount: true,
      },
    },
  );

  $: organizations = $organizationQuery.data?.organizations ?? [];
  $: projects = $projectsQuery.data?.projects ?? [];

  $: organizationPaths = {
    options: createOrgPaths(organizations, organization, planDisplayName),
  };

  function createOrgPaths(
    organizations: V1Organization[],
    viewingOrg: string | undefined,
    planDisplayName: string,
  ) {
    const pathMap = new Map<string, PathOption>();

    organizations.forEach(({ name, displayName }) => {
      pathMap.set(name.toLowerCase(), {
        label: displayName || name,
        pill: planDisplayName,
      });
    });

    if (!viewingOrg) return pathMap;

    if (!pathMap.has(viewingOrg.toLowerCase())) {
      pathMap.set(viewingOrg.toLowerCase(), {
        label: viewingOrg,
        pill: planDisplayName,
      });
    }

    return pathMap;
  }

  $: projectPaths = {
    options: projects.reduce(
      (map, { name }) =>
        map.set(name.toLowerCase(), { label: name, preloadData: false }),
      new Map<string, PathOption>(),
    ),
  };

  $: pathParts = [organizationPaths, projectPaths];
  $: currentPath = [organization, project];
</script>

<Header borderBottom={!onOrgPage}>
  <HeaderLogo href={rillLogoHref} logoUrl={organizationLogoUrl} />
  {#if organization}
    <Breadcrumbs {pathParts} {currentPath} />
  {/if}

  <div class="flex gap-x-2 items-center ml-auto">
    {#if $user.isSuccess}
      {#if $user.data?.user}
        <AvatarButton />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</Header>
