<script>
  import { page } from "$app/stores";
  import Home from "@rilldata/web-common/components/icons/Home.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import SignIn from "../authentication/SignIn.svelte";
  import UserButton from "../authentication/UserButton.svelte";
  import DeploymentStatusChip from "../deployments/DeploymentStatusChip.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";

  $: project = $page.params.project;

  const userQuery = createAdminServiceGetCurrentUser();
  $: signedIn = !!$userQuery.data?.user;
</script>

<div class="border-b flex items-center">
  <a
    href="/"
    class="inline-flex items-center py-2 px-3 hover:bg-gray-200"
    style="height:44px;"
  >
    <Tooltip distance={12}>
      <Home size="1.5em" color="black" />
      <TooltipContent slot="tooltip-content">Home</TooltipContent>
    </Tooltip>
  </a>
  <Breadcrumbs />
  {#if project}
    <div class="ml-3">
      <DeploymentStatusChip />
    </div>
  {/if}
  <div class="flex-grow" />
  <div class="p-2">
    {#if signedIn}
      <UserButton />
    {:else}
      <SignIn />
    {/if}
  </div>
</div>
