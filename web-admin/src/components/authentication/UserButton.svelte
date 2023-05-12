<script lang="ts">
  import SimpleActionMenu from "@rilldata/web-common/components/menu/wrappers/SimpleActionMenu.svelte";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import { ADMIN_URL } from "../../client/http-client";

  const user = createAdminServiceGetCurrentUser();

  function handleLogOut() {
    window.location.href = `${ADMIN_URL}/auth/logout?redirect=${window.location.origin}${window.location.pathname}`;
  }

  function handleDocumentation() {
    window.open("https://docs.rilldata.com", "_blank");
  }

  const isDev = process.env.NODE_ENV === "development";
</script>

<SimpleActionMenu
  options={[
    { main: "Logout", callback: handleLogOut },
    { main: "Documentation", callback: handleDocumentation },
  ]}
  let:toggleMenu
  minWidth="0px"
  distance={4}
>
  <img
    src={$user.data?.user?.photoUrl}
    alt="avatar"
    class="h-7 rounded-full cursor-pointer"
    referrerpolicy={isDev ? "no-referrer" : ""}
    on:click={toggleMenu}
    on:keydown={toggleMenu}
  />
</SimpleActionMenu>
