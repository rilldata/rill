<script lang="ts">
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomColor } from "@rilldata/web-common/features/themes/color-config";

  export let name: string;
  export let email: string;
  export let photoUrl: string | null = null;
  export let isCurrentUser: boolean = false;
  export let pendingAcceptance: boolean = false;
</script>

<div class="flex items-center gap-2">
  <Avatar
    size="h-7 w-7"
    src={photoUrl}
    alt={pendingAcceptance ? null : name}
    bgColor={getRandomColor(email)}
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
