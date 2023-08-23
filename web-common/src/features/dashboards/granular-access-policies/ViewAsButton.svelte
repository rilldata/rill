<script lang="ts">
  import { goto } from "$app/navigation";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Add from "../../../components/icons/Add.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import Check from "../../../components/icons/Check.svelte";
  import EyeIcon from "../../../components/icons/EyeIcon.svelte";
  import Spacer from "../../../components/icons/Spacer.svelte";
  import { Divider, Menu, MenuItem } from "../../../components/menu";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useTestUsers } from "./useTestUsers";
  import { viewAs, viewAsStore } from "./viewAs";

  export let dashboardName: string;

  const queryClient = useQueryClient();

  let viewAsMenuOpen = false;

  $: testUsers = useTestUsers($runtime.instanceId);

  const iconColor = "#15141A";
</script>

<WithTogglableFloatingElement
  alignment="start"
  distance={8}
  let:toggleFloatingElement
  location="bottom"
  on:open={() => (viewAsMenuOpen = true)}
  on:close={() => (viewAsMenuOpen = false)}
>
  <button
    class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 items-center"
    on:click={(evt) => {
      evt.stopPropagation();
      toggleFloatingElement();
    }}
  >
    <div class={$viewAsStore !== null && "text-blue-600"}>
      <EyeIcon size={"16px"} />
    </div>
    {#if $viewAsStore == null}
      <div class="flex items-center gap-x-1">
        <span>View as</span><CaretDownIcon />
      </div>
    {:else}
      <div class="text-blue-600">
        Viewing as <span class="font-bold">{$viewAsStore.email}</span>
      </div>
    {/if}
  </button>
  <Menu
    minWidth=""
    on:click-outside={toggleFloatingElement}
    on:escape={toggleFloatingElement}
    slot="floating-element"
  >
    <MenuItem
      icon
      selected={$viewAsStore === null}
      on:select={() => {
        toggleFloatingElement();
        viewAs(queryClient, dashboardName, null);
      }}
    >
      <svelte:fragment slot="icon">
        {#if $viewAsStore === null}
          <Check size="16px" color={iconColor} />
        {:else}
          <Spacer size="16px" />
        {/if}
      </svelte:fragment>
      Me
    </MenuItem>
    {#if $testUsers.data?.length > 0}
      <Divider />
      {#each $testUsers.data as user}
        <MenuItem
          icon
          selected={$viewAsStore?.email === user?.email}
          on:select={() => {
            toggleFloatingElement();
            viewAs(queryClient, dashboardName, user);
          }}
        >
          <svelte:fragment slot="icon">
            {#if $viewAsStore?.email === user?.email}
              <Check size="16px" color={iconColor} />
            {:else}
              <Spacer size="16px" />
            {/if}
          </svelte:fragment>
          {user.email}
        </MenuItem>
      {/each}
    {/if}
    <Divider />
    <MenuItem
      icon
      on:select={() => {
        toggleFloatingElement();
        goto(`/rill.yaml?addTestUser=true`);
      }}
    >
      <Add size="16px" slot="icon" color={iconColor} />
      <span>Add test user</span>
    </MenuItem>
  </Menu>
</WithTogglableFloatingElement>
