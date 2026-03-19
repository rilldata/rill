<script lang="ts">
  import { page } from "$app/stores";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListSuperusers,
  } from "../../client";
  import {
    useBreadcrumbOrgPaths,
    useBreadcrumbProjectPaths,
  } from "../navigation/breadcrumb-selectors";
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

  // Check if the current user is a superuser; the ListSuperusers call returns 403 for non-superusers
  const superusers = createAdminServiceListSuperusers();
  $: isSuperuser =
    $superusers.isSuccess &&
    !!$user.data?.user?.email &&
    ($superusers.data?.users ?? []).some(
      (su) => su.email === $user.data?.user?.email,
    );

  $: orgPathsQuery = useBreadcrumbOrgPaths(
    loggedIn,
    organization,
    planDisplayName,
  );
  $: projectPathsQuery = useBreadcrumbProjectPaths(organization, readProjects);

  $: pathParts = [
    { options: $orgPathsQuery.data ?? new Map() },
    { options: $projectPathsQuery.data ?? new Map() },
  ];
  $: currentPath = [organization, project];
</script>

<Header borderBottom={!onOrgPage}>
  <HeaderLogo href={rillLogoHref} logoUrl={organizationLogoUrl} />
  {#if organization}
    <Breadcrumbs {pathParts} {currentPath} />
  {/if}

  <div class="flex gap-x-2 items-center ml-auto">
    {#if isSuperuser}
      <a
        href="/-/admin"
        class="text-xs font-medium text-gray-500 hover:text-gray-700"
      >
        Admin
      </a>
    {/if}
    {#if $user.isSuccess}
      {#if $user.data?.user}
        <AvatarButton />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</Header>
