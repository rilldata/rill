<script lang="ts">
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { updateDevJWT } from "@rilldata/web-common/features/dashboards/granular-access-policies/updateDevJWT";
  import { IconSpaceFixer } from "../../../components/button";
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
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useProjectParser } from "../../entity-management/resource-selectors";

  const iconColor = "#15141A";

  let viewAsMenuOpen = false;
  let open = false;

  $: ({ instanceId } = $runtime);

  $: mockUsers = useMockUsers(instanceId);

  $: projectParserQuery = useProjectParser(queryClient, instanceId);

  $: showErrorBanner = $projectParserQuery.error?.response?.status === 404;
  $: if (showErrorBanner) {
    if ($projectParserQuery.error?.response?.data?.message) {
      eventBus.emit("banner", {
        type: "error",
        message: $projectParserQuery.error?.response?.data?.message,
        iconType: "alert",
      });
    }
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    {#if $selectedMockUserStore === null}
      <button
        use:builder.action
        {...builder}
        class="px-3 py-1.5 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 items-center"
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
        on:remove={() => {
          updateDevJWT(queryClient, null);
          eventBus.emit("banner", null);
        }}
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
  </DropdownMenu.Trigger>

  <DropdownMenu.Content>
    {#if !$mockUsers.data || $mockUsers.data?.length === 0}
      <DropdownMenu.Item disabled>No mock users</DropdownMenu.Item>
    {:else if $mockUsers.data?.length > 0}
      {#each $mockUsers.data as user}
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
