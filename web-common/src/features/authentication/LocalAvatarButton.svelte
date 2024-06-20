<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    createLocalServiceDeployValidation,
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
  } from "@rilldata/web-common/runtime-client/local-service";

  $: user = createLocalServiceGetCurrentUser();
  $: metadata = createLocalServiceGetMetadata();
  $: deployValidation = createLocalServiceDeployValidation();

  // $: if ($user.data?.authenticated) initPylonChat($user.data);

  function handlePylon() {
    // TODO
    // window.Pylon("show");
  }

  function handleSignIn() {
    window.location.href = `${$deployValidation.data?.loginUrl}/?redirect=${window.location.origin}${window.location.pathname}`;
  }

  function makeLogOutHref(): string {
    if (!$metadata.data?.adminUrl) return "";
    // Create a login URL that redirects back to the current page
    const loginWithRedirect = `${$metadata.data?.adminUrl}/auth/login?redirect=${window.location.origin}${window.location.pathname}`;

    // Create the logout URL, providing the login URL as a redirect
    return `${$metadata.data?.adminUrl}/auth/logout?redirect=${loginWithRedirect}`;
  }
</script>

{#if $user.isSuccess}
  {#if $user.data?.authenticated}
    <DropdownMenu.Root>
      <DropdownMenu.Trigger class="flex-none">
        <img
          src={$user.data?.photoUrl}
          alt="avatar"
          class="h-7 inline-flex items-center rounded-full"
          referrerpolicy="no-referrer"
        />
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
        <DropdownMenu.Item on:click={handlePylon}>
          Contact Rill support
        </DropdownMenu.Item>
        <DropdownMenu.Item
          href={makeLogOutHref()}
          class="text-gray-800 font-normal"
        >
          Logout
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {:else}
    <Button type="primary" on:click={handleSignIn}>Log In / Sign Up</Button>
  {/if}
{/if}
