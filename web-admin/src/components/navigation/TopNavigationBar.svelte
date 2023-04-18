<script>
  import { page } from "$app/stores";
  import Home from "@rilldata/web-common/components/icons/Home.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import SignIn from "../authentication/SignIn.svelte";
  import UserButton from "../authentication/UserButton.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";

  $: organization = $page.params.organization;

  const user = createAdminServiceGetCurrentUser({
    query: { placeholderData: undefined },
  });
</script>

<div
  class="border-b grid items-center w-full justify-stretch pr-4"
  style:grid-template-columns="max-content auto max-content"
>
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
  {#if organization}
    <Breadcrumbs />
  {:else}
    <div />
  {/if}
  <div class="flex gap-x-3 items-center">
    <a
      class="font-medium"
      href="https://discord.com/invite/ngVV4KzEGv?utm_source=rill&utm_medium=rill-cloud-nav"
      >Ask for help</a
    >
    {#if $user.isSuccess}
      <div>
        {#if $user.data && $user.data.user}
          <UserButton />
        {:else}
          <SignIn />
        {/if}
      </div>
    {/if}
  </div>
</div>
