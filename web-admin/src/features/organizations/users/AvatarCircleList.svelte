<script lang="ts">
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";

  export let name: string;
  export let email: string;
  export let isCurrentUser: boolean = false;
  export let pendingAcceptance: boolean = false;
</script>

<div class="flex items-center gap-2 py-2 pl-2">
  <Avatar
    size="h-7 w-7"
    alt={pendingAcceptance ? null : name}
    bgColor={getRandomBgColor()}
  />
  <div class="flex flex-col text-left">
    <span class="text-sm font-medium text-gray-900">
      {name}
      <span class="text-gray-500 font-normal">
        {isCurrentUser ? "(You)" : ""}
      </span>
    </span>
    <span class="text-xs text-gray-500"
      >{pendingAcceptance ? "Pending invitation" : email}</span
    >
  </div>
</div>
