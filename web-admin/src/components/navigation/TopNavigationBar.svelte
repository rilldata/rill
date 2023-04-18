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
  <Tooltip distance={2}>
    <a
      href="/"
      class="inline-flex items-center hover:bg-gray-200 grid place-items-center rounded"
      style:margin-left="8px"
      style:margin-top="4px"
      style:margin-bottom="4px"
      style:height="36px"
      style:width="36px"
    >
      <Home size="20px" color="black" />
    </a>
    <TooltipContent slot="tooltip-content">Home</TooltipContent>
  </Tooltip>
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
