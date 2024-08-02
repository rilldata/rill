<script lang="ts">
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import NoUser from "@rilldata/web-common/components/icons/NoUser.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    createLocalServiceDeployValidation,
    createLocalServiceGetCurrentUser,
  } from "@rilldata/web-common/runtime-client/local-service";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: user = createLocalServiceGetCurrentUser();
  $: deployValidation = createLocalServiceDeployValidation();

  $: loginUrl = `${$deployValidation.data?.loginUrl}/?redirect=${window.location.origin}${window.location.pathname}`;
  $: logoutUrl = `${$deployValidation.data?.loginUrl}/logout?redirect=${$page.url.href}`;
  $: loggedIn = $user.isSuccess && $user.data?.user;
</script>

{#if ($user.isLoading || $deployValidation.isLoading) && !$user.error && !$deployValidation.error}
  <div class="flex flex-row items-center h-7 mx-1.5">
    <Spinner size="16px" status={EntityStatus.Running} />
  </div>
{:else if $user.data && $deployValidation.data}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger class="flex-none w-7">
      {#if loggedIn}
        <img
          src={$user.data?.user?.photoUrl}
          alt="avatar"
          class="h-7 inline-flex items-center rounded-full"
          referrerpolicy="no-referrer"
        />
      {:else}
        <NoUser />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content>
      <DropdownMenu.Item
        href="https://docs.rilldata.com"
        target="_blank"
        rel="noreferrer noopener"
        class="text-gray-800 font-normal"
      >
        Documentation
      </DropdownMenu.Item>
      <DropdownMenu.Item
        href="https://discord.com/invite/ngVV4KzEGv?utm_source=rill&utm_medium=rill-cloud-avatar-menu"
        target="_blank"
        rel="noreferrer noopener"
        class="text-gray-800 font-normal"
      >
        Join us on Discord
      </DropdownMenu.Item>
      <!-- TODO -->
      <!-- <DropdownMenu.Item on:click={handlePylon}>-->
      <!--   Contact Rill support-->
      <!-- </DropdownMenu.Item>-->
      {#if loggedIn}
        <DropdownMenu.Item href={logoutUrl} class="text-gray-800 font-normal">
          Logout
        </DropdownMenu.Item>
      {:else}
        <DropdownMenu.Item href={loginUrl} class="text-gray-800 font-normal">
          Log in / Sign up
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
