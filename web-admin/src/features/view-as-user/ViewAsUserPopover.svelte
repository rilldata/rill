<script lang="ts">
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { Search } from "@rilldata/web-common/components/search";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { matchSorter } from "match-sorter";
  import { createEventDispatcher } from "svelte";
  import { createAdminServiceSearchProjectUsers, V1User } from "../../client";
  import { errorStore } from "../../components/errors/error-store";
  import { setViewedAsUser } from "./setViewedAsUser";
  import { viewAsUserStore } from "./viewAsUserStore";

  export let organization: string;
  export let project: string;

  // Note: this approach will break down if/when there are more than 1000 users in a project
  $: projectUsers = createAdminServiceSearchProjectUsers(
    organization,
    project,
    { emailQuery: "%", pageSize: 1000, pageToken: undefined }
  );

  const dispatch = createEventDispatcher();

  const queryClient = useQueryClient();
  async function handleViewAsUser(user: V1User) {
    await setViewedAsUser(queryClient, organization, project, user);
    errorStore.reset();
    dispatch("select");
  }

  let minWidth = "150px";
  let maxWidth = "300px";
  let minHeight = "150px";
  let maxHeight = "190px";

  let searchText = "";

  $: clientSideUsers = $projectUsers.data?.users ?? [];

  $: visibleUsers = searchText
    ? matchSorter(clientSideUsers, searchText, { keys: ["email"] })
    : clientSideUsers;
</script>

<Menu
  focusOnMount={false}
  {minWidth}
  {maxWidth}
  {minHeight}
  {maxHeight}
  paddingBottom={0}
  paddingTop={1}
  rounded={false}
  on:click-outside
  on:escape
>
  <div class="px-2 pt-1 pb-2 text-[10px] text-gray-500 text-left">
    Test your <a
      target="_blank"
      href="https://docs.rilldata.com/develop/security">security policies</a
    > by viewing this project from the perspective of another user.
  </div>
  <Search bind:value={searchText} />
  {#if visibleUsers.length > 0}
    <div class="overflow-auto pb-1">
      {#each visibleUsers as user}
        <MenuItem
          icon
          animateSelect={false}
          focusOnMount={false}
          on:select={() => handleViewAsUser(user)}
        >
          <svelte:fragment slot="icon">
            {#if user === $viewAsUserStore}
              <Check size="20px" color="#15141A" />
            {:else}
              <Spacer size="20px" />
            {/if}
          </svelte:fragment>
          {user.email}
        </MenuItem>
      {/each}
    </div>
  {:else}
    <div class="mt-5 ui-copy-disabled text-center">no results</div>
  {/if}
</Menu>
