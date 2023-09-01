<script lang="ts">
  import { page } from "$app/stores";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import Menu from "@rilldata/web-common/components/menu/core/Menu.svelte";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import { ADMIN_URL } from "../../client/http-client";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";

  const user = createAdminServiceGetCurrentUser();

  let menuOpen = false;

  function handleViewAs() {
    console.log("open 'view as' UI");
  }

  function handleLogOut() {
    const loginWithRedirect = `${ADMIN_URL}/auth/login?redirect=${window.location.origin}${window.location.pathname}`;
    window.location.href = `${ADMIN_URL}/auth/logout?redirect=${loginWithRedirect}`;
  }

  function handleDocumentation() {
    window.open("https://docs.rilldata.com", "_blank");
  }

  const isDev = process.env.NODE_ENV === "development";
</script>

<WithTogglableFloatingElement
  distance={4}
  location="bottom"
  alignment="start"
  let:handleClose
  let:toggleFloatingElement={toggleMenu}
  on:open={() => (menuOpen = true)}
  on:close={() => (menuOpen = false)}
>
  <img
    src={$user.data?.user?.photoUrl}
    alt="avatar"
    class="h-7 rounded-full cursor-pointer"
    referrerpolicy={isDev ? "no-referrer" : ""}
    on:click={toggleMenu}
    on:keydown={toggleMenu}
  />
  <Menu
    slot="floating-element"
    minWidth="0px"
    focusOnMount={false}
    on:select-item={handleClose}
    on:click-outside={handleClose}
    on:escape={handleClose}
  >
    {#if $page.params.organization && $page.params.project}
      <ProjectAccessControls
        organization={$page.params.organization}
        project={$page.params.project}
      >
        <MenuItem
          slot="manage-project"
          on:select={() => {
            handleClose();
            handleViewAs();
          }}
        >
          View as
        </MenuItem>
      </ProjectAccessControls>
    {/if}
    <MenuItem
      on:select={() => {
        handleClose();
        handleLogOut();
      }}>Logout</MenuItem
    >
    <MenuItem
      on:select={() => {
        handleClose();
        handleDocumentation();
      }}>Documentation</MenuItem
    >
  </Menu>
</WithTogglableFloatingElement>
