<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  export let name: string;
  export let email: string | null = null;
  export let photoUrl: string | null = null;
  export let isCurrentUser: boolean = false;
  export let pendingAcceptance: boolean = false;
  export let shape: "circle" | "square" = "circle";
  export let count: number = 0;
  export let isEveryoneFromText: boolean = false;
  export let canManage: boolean = false;

  $: organization = $page.params.organization;

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
</script>

<div class="flex items-center gap-2 py-2 pl-2">
  {#if shape === "circle"}
    <Avatar
      avatarSize="h-7 w-7"
      src={photoUrl}
      alt={pendingAcceptance ? null : name}
      bgColor={getRandomBgColor(email ?? name)}
    />
  {:else if shape === "square"}
    <div
      class={cn(
        "h-7 w-7 rounded-sm flex items-center justify-center",
        getRandomBgColor(email ?? name),
      )}
    >
      <span class="text-sm text-white font-semibold">{getInitials(name)}</span>
    </div>
  {/if}
  <div class="flex flex-col text-left">
    <span class="text-sm font-medium text-gray-900">
      {#if isEveryoneFromText}
        {@html `Everyone from <span class="font-bold">${name}</span>`}
      {:else}
        {name}
      {/if}
      <span class="text-gray-500 font-normal">
        {isCurrentUser ? "(You)" : ""}
      </span>
    </span>
    {#if pendingAcceptance || email}
      <span class="text-xs text-gray-500">
        {pendingAcceptance ? "Pending invitation" : email}
      </span>
    {/if}
    <div class="flex flex-row items-center gap-x-1">
      {#if count && count > 0}
        <span class="text-xs text-gray-500">
          {count} user{count > 1 ? "s" : ""}
        </span>
      {/if}
      {#if canManage}
        <!-- FIXME: onClick https://www.figma.com/design/Qt6EyotCBS3V6O31jVhMQ7/RILL-Latest?node-id=17077-830426&node-type=frame&t=VQugAVqX0LGPrjVJ-0 -->
        <button class="text-xs text-primary-600 font-medium cursor-pointer">
          Manage
        </button>
      {/if}
    </div>
  </div>
</div>
