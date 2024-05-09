<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import { ADMIN_URL } from "../../client/http-client";
  import { initPylonChat } from "../help/initPylonChat";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";
  import ViewAsUserPopover from "../view-as-user/ViewAsUserPopover.svelte";

  const isDev = process.env.NODE_ENV === "development";
  const user = createAdminServiceGetCurrentUser();

  let subMenuOpen = false;

  $: if ($user.data?.user) initPylonChat($user.data.user);

  function handleDocumentation() {
    window.open("https://docs.rilldata.com", "_blank");
  }

  function handleDiscord() {
    window.open(
      "https://discord.com/invite/ngVV4KzEGv?utm_source=rill&utm_medium=rill-cloud-avatar-menu",
      "_blank",
    );
  }

  function handlePylon() {
    window.Pylon("show");
  }

  function handleLogOut() {
    // Create a login URL that redirects back to the current page
    const loginWithRedirect = `${ADMIN_URL}/auth/login?redirect=${window.location.origin}${window.location.pathname}`;

    // Go to the logout URL, providing the login URL as a redirect
    window.location.href = `${ADMIN_URL}/auth/logout?redirect=${loginWithRedirect}`;
  }

  function handleAlerts() {
    goto(`/${$page.params.organization}/${$page.params.project}/-/alerts`);
  }

  function handleReports() {
    goto(`/${$page.params.organization}/${$page.params.project}/-/reports`);
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="flex-none">
    <img
      src={$user.data?.user?.photoUrl}
      alt="avatar"
      class="h-7 inline-flex items-center rounded-full cursor-pointer"
      referrerpolicy={isDev ? "no-referrer" : ""}
    />
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    {#if $page.params.organization && $page.params.project && $page.params.dashboard}
      <ProjectAccessControls
        organization={$page.params.organization}
        project={$page.params.project}
      >
        <svelte:fragment slot="manage-project">
          <DropdownMenu.Sub bind:open={subMenuOpen}>
            <DropdownMenu.SubTrigger
              disabled={true}
              on:click={() => (subMenuOpen = !subMenuOpen)}
            >
              View as
            </DropdownMenu.SubTrigger>
            <DropdownMenu.SubContent
              class="flex flex-col min-w-[150px] max-w-[300px] min-h-[150px] max-h-[190px]"
            >
              <ViewAsUserPopover
                organization={$page.params.organization}
                project={$page.params.project}
              />
            </DropdownMenu.SubContent>
          </DropdownMenu.Sub>
        </svelte:fragment>
      </ProjectAccessControls>
      <DropdownMenu.Item on:click={handleAlerts}>Alerts</DropdownMenu.Item>
      <DropdownMenu.Item on:click={handleReports}>Reports</DropdownMenu.Item>
    {/if}
    <DropdownMenu.Item on:click={handleDocumentation}>
      Documentation
    </DropdownMenu.Item>
    <DropdownMenu.Item on:click={handleDiscord}>
      Join us on Discord
    </DropdownMenu.Item>
    <DropdownMenu.Item on:click={handlePylon}>
      Contact Rill support
    </DropdownMenu.Item>
    <DropdownMenu.Item on:click={handleLogOut}>Logout</DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
