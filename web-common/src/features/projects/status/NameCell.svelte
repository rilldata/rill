<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { User } from "lucide-svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let {
    name,
    isPersonal = false,
    currentUserId,
    ownerId = undefined,
  }: {
    name: string;
    isPersonal?: boolean;
    currentUserId?: string | undefined;
    ownerId?: string | undefined;
  } = $props();

  let isCurrentUser = $derived(
    !!currentUserId && !!ownerId && currentUserId === ownerId,
  );

  // There is no good API to get a user's info by id. GetUser API is for super users only.
  // TODO: if we want to show the exact owner then we either need to open up that API or hit ListProjectMemberUsers.
</script>

<div class="flex flex-row items-center truncate">
  {#if isPersonal}
    <Tooltip.Root>
      <Tooltip.Trigger>
        <User class="mr-1" size={14} />
      </Tooltip.Trigger>
      <Tooltip.Content side="bottom" sideOffset={8}>
        {isCurrentUser ? m.status_owned_by_you() : m.status_owned_by_other()}
      </Tooltip.Content>
    </Tooltip.Root>
  {/if}
  <div>{name}</div>
</div>
