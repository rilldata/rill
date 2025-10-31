<script lang="ts">
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { updateDevJWT } from "@rilldata/web-common/features/dashboards/granular-access-policies/updateDevJWT";

  import { Chip } from "../../../components/chip";
  import Add from "../../../components/icons/Add.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import Check from "../../../components/icons/Check.svelte";
  import EyeIcon from "../../../components/icons/EyeIcon.svelte";
  import Spacer from "../../../components/icons/Spacer.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { selectedMockUserStore } from "./stores";
  import { useMockUsers } from "./useMockUsers";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  const iconColor = "#15141A";

  let viewAsMenuOpen = false;
  let open = false;

  $: ({ instanceId } = $runtime);

  $: mockUsers = useMockUsers(instanceId);
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    {#if $selectedMockUserStore === null}
      <button
        use:builder.action
        {...builder}
        class="px-3 py-1.5 rounded flex flex-row gap-x-2 hover:bg-gray-200 items-center"
      >
        <EyeIcon size={"16px"} />
        <div class="flex items-center gap-x-1">
          <span>View as</span><CaretDownIcon />
        </div>
      </button>
    {:else}
      <Chip
        builders={[builder]}
        removable
        slideDuration={0}
        active={viewAsMenuOpen}
        removeTooltipText="Clear view"
        onRemove={() => {
          updateDevJWT(queryClient, null);
        }}
      >
        <div slot="body">
          Viewing as <b>{$selectedMockUserStore.email}</b>
        </div>
      </Chip>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    {#if !$mockUsers.data || $mockUsers.data?.length === 0}
      <DropdownMenu.Item disabled>No mock users</DropdownMenu.Item>
    {:else if $mockUsers.data?.length > 0}
      {#each $mockUsers.data as user (user?.email)}
        <DropdownMenu.Item
          on:click={() => {
            updateDevJWT(queryClient, user);
          }}
          class="flex gap-x-2 items-center"
        >
          {#if $selectedMockUserStore?.email === user?.email}
            <Check size="16px" color={iconColor} />
          {:else}
            <Spacer size="16px" />
          {/if}

          {user.email}
        </DropdownMenu.Item>
      {/each}
    {/if}
    <DropdownMenu.Separator />
    <DropdownMenu.Item
      href={`/files/rill.yaml?addMockUser=true`}
      class="flex gap-x-2 items-center text-black font-normal"
    >
      <Add color={iconColor} size="16px" />
      Add mock user
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
