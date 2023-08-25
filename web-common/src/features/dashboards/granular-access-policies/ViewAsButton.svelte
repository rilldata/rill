<script lang="ts">
  import { goto } from "$app/navigation";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { IconSpaceFixer } from "../../../components/button";
  import { Chip } from "../../../components/chip";
  import Add from "../../../components/icons/Add.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import Check from "../../../components/icons/Check.svelte";
  import EyeIcon from "../../../components/icons/EyeIcon.svelte";
  import Spacer from "../../../components/icons/Spacer.svelte";
  import { Divider, Menu, MenuItem } from "../../../components/menu";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { selectedMockUserStore } from "./stores";
  import { useMockUsers } from "./useMockUsers";

  let viewAsMenuOpen = false;

  $: mockUsers = useMockUsers($runtime.instanceId);

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
  {#if $selectedMockUserStore === null}
    <button
      class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 items-center"
      on:click={toggleFloatingElement}
    >
      <EyeIcon size={"16px"} />
      <div class="flex items-center gap-x-1">
        <span>View as</span><CaretDownIcon />
      </div>
    </button>
  {:else}
    <Chip
      removable
      on:click={toggleFloatingElement}
      on:remove={() => {
        if (viewAsMenuOpen) toggleFloatingElement();
        selectedMockUserStore.set(null);
      }}
      active={viewAsMenuOpen}
    >
      <div slot="body" class="flex gap-x-2">
        <div>
          Viewing as <span class="font-bold"
            >{$selectedMockUserStore.email}</span
          >
        </div>
        <div class="flex items-center">
          <IconSpaceFixer pullRight>
            <div
              class="transition-transform"
              class:-rotate-180={viewAsMenuOpen}
            >
              <CaretDownIcon size="14px" />
            </div>
          </IconSpaceFixer>
        </div>
      </div>
      <svelte:fragment slot="remove-tooltip">
        <slot name="remove-tooltip-content">Clear view</slot>
      </svelte:fragment>
    </Chip>
  {/if}
  <Menu
    minWidth=""
    on:click-outside={toggleFloatingElement}
    on:escape={toggleFloatingElement}
    slot="floating-element"
  >
    {#if $mockUsers.data?.length > 0}
      {#each $mockUsers.data as user}
        <MenuItem
          icon
          selected={$selectedMockUserStore?.email === user?.email}
          on:select={() => {
            toggleFloatingElement();
            selectedMockUserStore.set(user);
          }}
        >
          <svelte:fragment slot="icon">
            {#if $selectedMockUserStore?.email === user?.email}
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
        goto(`/rill.yaml?addMockUser=true`);
      }}
    >
      <Add size="16px" slot="icon" color={iconColor} />
      <span>Add mock user</span>
    </MenuItem>
  </Menu>
</WithTogglableFloatingElement>
