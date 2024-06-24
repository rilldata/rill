<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import NoUser from "@rilldata/web-common/components/icons/NoUser.svelte";
  import {
    createLocalServiceDeployValidation,
    createLocalServiceGetCurrentUser,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: user = createLocalServiceGetCurrentUser();
  $: deployValidation = createLocalServiceDeployValidation();

  function handleSignIn() {
    window.location.href = `${$deployValidation.data?.loginUrl}/?redirect=${window.location.origin}${window.location.pathname}`;
  }

  function makeLogOutHref(): string {
    // Create the logout URL, providing the current URL
    return `${$runtime.host}/auth/logout?redirect=${$page.url.href}`;
  }

  $: loggedIn = $user.isSuccess && $user.data?.user;
</script>

{#if !$user.isFetching}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger class="flex-none">
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
        <DropdownMenu.Item
          href={makeLogOutHref()}
          class="text-gray-800 font-normal"
        >
          Logout
        </DropdownMenu.Item>
      {:else}
        <DropdownMenu.Item
          on:click={handleSignIn}
          class="text-gray-800 font-normal"
        >
          Log in / Sign up
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
