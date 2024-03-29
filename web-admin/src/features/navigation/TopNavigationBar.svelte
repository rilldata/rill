<script>
  import { page } from "$app/stores";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import Home from "@rilldata/web-common/components/icons/Home.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import ViewAsUserChip from "../../features/view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../../features/view-as-user/viewAsUserStore";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import SignIn from "../authentication/SignIn.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import ShareDashboardButton from "../dashboards/share/ShareDashboardButton.svelte";
  import { isErrorStoreEmpty } from "../errors/error-store";
  import ShareProjectButton from "../projects/ShareProjectButton.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";
  import { isDashboardPage, isProjectPage } from "./nav-utils";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import CreateAlert from "../alerts/CreateAlert.svelte";

  const user = createAdminServiceGetCurrentUser();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: dashboard = $page.params.dashboard;

  $: onProjectPage = isProjectPage($page);
  $: onDashboardPage = isDashboardPage($page);
</script>

<div
  class="grid items-center w-full justify-stretch pr-4 {onProjectPage
    ? ''
    : 'border-b'}"
  style:grid-template-columns="max-content auto max-content"
>
  <Tooltip distance={2}>
    <a
      class="inline-flex items-center hover:bg-gray-200 grid place-items-center rounded"
      href="/"
      style:height="36px"
      style:margin-bottom="4px"
      style:margin-left="8px"
      style:margin-top="4px"
      style:width="36px"
    >
      <Home color="black" size="20px" />
    </a>
    <TooltipContent slot="tooltip-content">Home</TooltipContent>
  </Tooltip>
  {#if $isErrorStoreEmpty && organization}
    <Breadcrumbs />
  {:else}
    <div />
  {/if}
  <div class="flex gap-x-2 items-center">
    {#if $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    {#if onProjectPage}
      <ShareProjectButton {organization} {project} />
    {/if}
    {#if onDashboardPage}
      <StateManagersProvider metricsViewName={dashboard}>
        <LastRefreshedDate {dashboard} />
        {#if $user.isSuccess && $user.data.user}
          <CreateAlert />
          <Bookmarks />
        {/if}
        <ShareDashboardButton />
      </StateManagersProvider>
    {/if}
    {#if $user.isSuccess}
      {#if $user.data && $user.data.user}
        <AvatarButton />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</div>
