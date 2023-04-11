<script>
  import { page } from "$app/stores";
  import RillLogo from "@rilldata/web-common/components/icons/RillLogo.svelte";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import SignIn from "../authentication/SignIn.svelte";
  import UserButton from "../authentication/UserButton.svelte";
  import DeploymentStatusChip from "../deployments/DeploymentStatusChip.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";

  $: project = $page.params.project;

  const userQuery = createAdminServiceGetCurrentUser();
  $: signedIn = !!$userQuery.data?.user;
</script>

<div class="border-b p-2 flex items-center">
  <a href="/" class="mr-3">
    <RillLogo iconOnly size={"2.25em"} />
  </a>
  <Breadcrumbs />
  {#if project}
    <div class="ml-3">
      <DeploymentStatusChip />
    </div>
  {/if}
  <div class="flex-grow" />
  {#if signedIn}
    <UserButton />
  {:else}
    <SignIn />
  {/if}
</div>
