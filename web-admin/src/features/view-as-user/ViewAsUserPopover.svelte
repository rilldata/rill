<script lang="ts">
  import * as Command from "@rilldata/web-common/components/command/index.js";
  import {
    createAdminServiceSearchProjectUsers,
    type V1User,
  } from "../../client";
  import { errorStore } from "../../components/errors/error-store";
  import { viewAsUserStore } from "./viewAsUserStore";

  export let organization: string;
  export let project: string;
  export let onSelectUser: (user: V1User) => void;

  // Note: this approach will break down if/when there are more than 1000 users in a project
  $: projectUsers = createAdminServiceSearchProjectUsers(
    organization,
    project,
    { emailQuery: "%", pageSize: 1000, pageToken: undefined },
  );

  function handleViewAsUser(user: V1User) {
    viewAsUserStore.set(user);
    errorStore.reset();
    onSelectUser(user);
  }

  $: clientSideUsers = $projectUsers.data?.users ?? [];
</script>

<div class="px-0.5 pt-1 pb-2 text-[10px] text-gray-500 text-left">
  Test your <a
    target="_blank"
    href="https://docs.rilldata.com/build/metrics-view/security#rill-cloud"
    >security policies</a
  > by viewing this project from the perspective of another user.
</div>

<Command.Root>
  <Command.Input placeholder="Search for users" />
  <Command.List>
    <Command.Empty>No results found.</Command.Empty>
    <Command.Group heading="Users">
      {#each clientSideUsers as user}
        <Command.Item onSelect={() => handleViewAsUser(user)}>
          {user.email}
        </Command.Item>
      {/each}
    </Command.Group>
  </Command.List>
</Command.Root>
