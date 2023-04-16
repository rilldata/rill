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
  {#if organization}
    <Breadcrumbs />
  {/if}
  <div class="flex-grow" />
  {#if $user.isSuccess}
    <div class="p-2">
      {#if $user.data && $user.data.user}
        <UserButton />
      {:else}
        <SignIn />
      {/if}
    </div>
  {/if}
</div>
