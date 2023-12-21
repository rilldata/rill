<script lang="ts">
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import { ADMIN_URL } from "../../client/http-client";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";
  import ViewAsUserPopover from "../view-as-user/ViewAsUserPopover.svelte";

  const user = createAdminServiceGetCurrentUser();

  function handleDocumentation() {
    window.open("https://docs.rilldata.com", "_blank");
  }

  function handleAskForHelp() {
    window.open(
      "https://discord.com/invite/ngVV4KzEGv?utm_source=rill&utm_medium=rill-cloud-avatar-menu",
      "_blank"
    );
  }

  function handleLogOut() {
    const loginWithRedirect = `${ADMIN_URL}/auth/login?redirect=${window.location.origin}${window.location.pathname}`;
    window.location.href = `${ADMIN_URL}/auth/logout?redirect=${loginWithRedirect}`;
  }

  const isDev = process.env.NODE_ENV === "development";

  let subMenuOpen = false;
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger>
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
    {/if}
    <DropdownMenu.Item on:click={handleDocumentation}>
      Documentation
    </DropdownMenu.Item>
    <DropdownMenu.Item on:click={handleAskForHelp}>
      Ask for help
    </DropdownMenu.Item>
    <DropdownMenu.Item on:click={handleLogOut}>Logout</DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
