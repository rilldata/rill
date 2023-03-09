<script>
  import RillLogo from "@rilldata/web-common/components/icons/RillLogo.svelte";
  import { useAdminServiceGetCurrentUser } from "../../client";
  import SignIn from "../authentication/SignIn.svelte";
  import UserButton from "../authentication/UserButton.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";

  const userQuery = useAdminServiceGetCurrentUser();
  $: signedIn = !!$userQuery.data?.user;
</script>

<div class="border-b p-2 flex items-center">
  <a href="/" class="mr-3">
    <RillLogo iconOnly size={"2.25em"} />
  </a>
  <Breadcrumbs />
  <div class="flex-grow" />
  {#if signedIn}
    <UserButton />
  {:else}
    <SignIn />
  {/if}
</div>
