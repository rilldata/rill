<script context="module" lang="ts">
  // Navigating between org and project pages swaps OrgHeader for ProjectHeader
  // (and vice versa). Both headers render an AvatarButton, but the old one
  // unmounts before the new one mounts. A normal <img> would be destroyed and
  // recreated, forcing the browser to re-decode the photo — causing a visible
  // flicker or broken-image flash.
  //
  // To avoid this, we keep a single <img> element at module scope. Each
  // AvatarButton instance adopts it via appendChild on mount and detaches it
  // on unmount. The browser retains the decoded image data, so it paints
  // instantly with no flash.
  let sharedImg: HTMLImageElement | null = null;
</script>

<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { redirectToLogout } from "@rilldata/web-admin/client/redirect-utils";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    initPylonChat,
    type UserLike,
  } from "@rilldata/web-common/features/help/initPylonChat";
  import { posthogIdentify } from "@rilldata/web-common/lib/analytics/posthog";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListSuperusers,
  } from "../../client";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";
  import ViewAsUserPopover from "../view-as-user/ViewAsUserPopover.svelte";
  import ThemeToggle from "@rilldata/web-common/features/themes/ThemeToggle.svelte";

  const user = createAdminServiceGetCurrentUser();
  // Fire ListSuperusers once per session to check if the avatar menu should
  // show the "Superuser Console" link. Non-superusers get a single 403 that
  // TanStack Query silently caches as an error (retry: false, staleTime: Infinity
  // ensures no repeated requests across component remounts).
  const superusers = createAdminServiceListSuperusers({
    query: {
      enabled: !!$user.data?.user?.email,
      retry: false,
      staleTime: Infinity,
    },
  });
  $: isSuperuser =
    $superusers.isSuccess &&
    !!$user.data?.user?.email &&
    ($superusers.data?.users ?? []).some(
      (su) => su.email === $user.data?.user?.email,
    );

  let imgContainer: HTMLElement;
  let primaryMenuOpen = false;
  let subMenuOpen = false;

  onMount(() => {
    const photoUrl = $user.data?.user?.photoUrl;
    if (!sharedImg) {
      sharedImg = document.createElement("img");
      sharedImg.className = "h-7 w-7 rounded-full";
      sharedImg.referrerPolicy = "no-referrer";
      sharedImg.alt = "avatar";
    }
    if (photoUrl && sharedImg.src !== photoUrl) {
      sharedImg.src = photoUrl;
    }
    imgContainer.appendChild(sharedImg);
    return () => {
      // Only detach if we're still the owner; a newer instance may have
      // already adopted the element via appendChild.
      if (sharedImg?.parentNode === imgContainer) {
        sharedImg.remove();
      }
    };
  });

  // Keep src in sync if the user query resolves or changes after mount.
  // sharedImg is module-level (shared singleton); reactivity is driven by $user.data.
  // svelte-ignore reactive_declaration_module_script_dependency
  $: if (
    sharedImg &&
    $user.data?.user?.photoUrl &&
    sharedImg.src !== $user.data.user.photoUrl
  ) {
    sharedImg.src = $user.data.user.photoUrl;
  }

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
    <div bind:this={imgContainer} class="h-7 w-7"></div>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    {#if params.organization && params.project}
      <ProjectAccessControls
        organization={params.organization}
        project={params.project}
      >
        <svelte:fragment slot="manage-project">
          <DropdownMenu.Sub bind:open={subMenuOpen}>
            <DropdownMenu.SubTrigger
              onclick={() => {
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
      {#if params.dashboard}
        <DropdownMenu.Item
          href={`/${params.organization}/${params.project}/-/alerts`}
        >
          Alerts
        </DropdownMenu.Item>
        <DropdownMenu.Item
          href={`/${params.organization}/${params.project}/-/reports`}
        >
          Reports
        </DropdownMenu.Item>
      {/if}
    {/if}

    {#if isSuperuser}
      <DropdownMenu.Item href="/-/superuser"
        >Superuser Console</DropdownMenu.Item
      >
      <DropdownMenu.Separator />
    {/if}

    <ThemeToggle />
    <DropdownMenu.Separator />

    <DropdownMenu.Item
      href="https://docs.rilldata.com"
      target="_blank"
      rel="noreferrer noopener"
    >
      Documentation
    </DropdownMenu.Item>
    <DropdownMenu.Item
      href="https://discord.gg/2ubRfjC7Rh"
      target="_blank"
      rel="noreferrer noopener"
    >
      Join us on Discord
    </DropdownMenu.Item>
    <DropdownMenu.Item onclick={handlePylon}>
      Contact Rill support
    </DropdownMenu.Item>
    <DropdownMenu.Item onclick={redirectToLogout}>Logout</DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
