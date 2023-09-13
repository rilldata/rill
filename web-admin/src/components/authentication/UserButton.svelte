<script lang="ts">
  import { page } from "$app/stores";
  import {
    Popover,
    PopoverButton,
    PopoverPanel,
  } from "@rgossiaux/svelte-headlessui";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import Menu from "@rilldata/web-common/components/menu/core/Menu.svelte";
  import { createPopperActions } from "svelte-popperjs";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import { ADMIN_URL } from "../../client/http-client";
  import ViewAsUserMenuItem from "../../features/view-as-user/ViewAsUserMenuItem.svelte";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";

  const user = createAdminServiceGetCurrentUser();

  function handleLogOut() {
    const loginWithRedirect = `${ADMIN_URL}/auth/login?redirect=${window.location.origin}${window.location.pathname}`;
    window.location.href = `${ADMIN_URL}/auth/logout?redirect=${loginWithRedirect}`;
  }

  function handleDocumentation() {
    window.open("https://docs.rilldata.com", "_blank");
  }

  const isDev = process.env.NODE_ENV === "development";

  // Position the Menu popover
  const [popperRef1, popperContent1] = createPopperActions();
  const popperOptions1 = {
    placement: "bottom-end",
    strategy: "fixed",
    modifiers: [{ name: "offset", options: { offset: [0, 4] } }],
  };

  // Position the View As User popover
  const [popperRef2, popperContent2] = createPopperActions();
</script>

<Popover class="relative" let:close={close1}>
  <PopoverButton use={[popperRef1]}>
    <img
      src={$user.data?.user?.photoUrl}
      alt="avatar"
      class="h-7 rounded-full cursor-pointer"
      referrerpolicy={isDev ? "no-referrer" : ""}
    />
  </PopoverButton>
  <PopoverPanel
    use={[popperRef2, [popperContent1, popperOptions1]]}
    class="max-w-fit absolute z-[1000]"
  >
    <Menu minWidth="0px" focusOnMount={false}>
      {#if $page.params.organization && $page.params.project && $page.params.dashboard}
        <ProjectAccessControls
          organization={$page.params.organization}
          project={$page.params.project}
        >
          <svelte:fragment slot="manage-project">
            <ViewAsUserMenuItem
              popperContent={popperContent2}
              on:select-user={() => close1(undefined)}
            />
          </svelte:fragment>
        </ProjectAccessControls>
      {/if}

      <MenuItem
        on:select={() => {
          // handleClose();
          handleLogOut();
        }}>Logout</MenuItem
      >
      <MenuItem
        on:select={() => {
          // handleClose();
          handleDocumentation();
        }}>Documentation</MenuItem
      >
    </Menu>
  </PopoverPanel>
</Popover>
