<script lang="ts">
  import { page } from "$app/stores";
  import { redirectToLogout } from "@rilldata/web-admin/client/redirect-utils";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import {
    initPylonChat,
    type UserLike,
  } from "@rilldata/web-common/features/help/initPylonChat";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";
  import ViewAsUserPopover from "../view-as-user/ViewAsUserPopover.svelte";

  const user = createAdminServiceGetCurrentUser();

  let subMenuOpen = false;

  $: if ($user.data?.user) initPylonChat($user.data.user as UserLike);
  $: ({ params } = $page);

  function handlePylon() {
    window.Pylon("show");
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="flex-none">
    <img
      src={$user.data?.user?.photoUrl}
      alt="avatar"
      class="h-7 inline-flex items-center rounded-full"
      referrerpolicy="no-referrer"
    />
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    {#if params.organization && params.project && params.dashboard}
      <ProjectAccessControls
        organization={params.organization}
        project={params.project}
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
                organization={params.organization}
                project={params.project}
              />
            </DropdownMenu.SubContent>
          </DropdownMenu.Sub>
        </svelte:fragment>
      </ProjectAccessControls>
      <DropdownMenu.Item
        href={`/${params.organization}/${params.project}/-/alerts`}
        class="text-gray-800 font-normal"
      >
        Alerts
      </DropdownMenu.Item>
      <DropdownMenu.Item
        href={`/${params.organization}/${params.project}/-/reports`}
        class="text-gray-800 font-normal"
      >
        Reports
      </DropdownMenu.Item>
    {/if}
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
      on:click={redirectToLogout}
      class="text-gray-800 font-normal"
    >
      Logout
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
