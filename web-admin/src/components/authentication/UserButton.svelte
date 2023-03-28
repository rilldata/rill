<script lang="ts">
  import SimpleActionMenu from "@rilldata/web-common/components/menu/wrappers/SimpleActionMenu.svelte";
  import { useAdminServiceGetCurrentUser } from "../../client";
  import { ADMIN_URL } from "../../client/http-client";

  const user = useAdminServiceGetCurrentUser();

  function handleLogOut() {
    window.location.href = `${ADMIN_URL}/auth/logout?redirect=${window.origin}`;
  }

  const isDev = process.env.NODE_ENV === "development";
</script>

<SimpleActionMenu
  options={[{ main: "Logout", callback: handleLogOut }]}
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
