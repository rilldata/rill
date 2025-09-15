<script lang="ts">
  import { page } from "$app/stores";
  import { redirectToLogout } from "@rilldata/web-admin/client/redirect-utils";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    initPylonChat,
    type UserLike,
  } from "@rilldata/web-common/features/help/initPylonChat";
  import { posthogIdentify } from "@rilldata/web-common/lib/analytics/posthog";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";
  import ViewAsUserPopover from "../view-as-user/ViewAsUserPopover.svelte";

  const user = createAdminServiceGetCurrentUser();

  let primaryMenuOpen = false;
  let subMenuOpen = false;

  $: if ($user.data?.user) {
    // Actions to take when the user is known
    posthogIdentify($user.data.user.id, {
      email: $user.data.user.email,
    });
    initPylonChat($user.data.user as UserLike);
  }

  $: ({ params } = $page);

  function handlePylon() {
    window.Pylon("show");
  }
</script>

<DropdownMenu.Root bind:open={primaryMenuOpen}>
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
              on:click={() => {
                subMenuOpen = !subMenuOpen;
              }}
            >
              View as
            </DropdownMenu.SubTrigger>
            <DropdownMenu.SubContent
              class="flex flex-col min-w-[150px] max-w-[300px]"
            >
              <ViewAsUserPopover
                organization={params.organization}
                project={params.project}
                onSelectUser={() => {
                  subMenuOpen = false;
                  primaryMenuOpen = false;
                }}
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
      href="https://discord.gg/2ubRfjC7Rh"
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
