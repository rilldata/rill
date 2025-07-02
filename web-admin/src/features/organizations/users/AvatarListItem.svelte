<script lang="ts">
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { page } from "$app/stores";

  export let name: string;
  export let email: string | null = null;
  export let photoUrl: string | null = null;
  export let isCurrentUser: boolean = false;
  export let pendingAcceptance: boolean = false;
  export let shape: "circle" | "square" = "circle";
  export let count: number = 0;
  export let role: string | null = null;
  export let leftSpacing: boolean = true;
  export let showGuestChip: boolean = false;
  export let showManage: boolean = false;

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }

  function handleManageClick() {
    const organization = $page.params.organization;
    window.open(
      `/${organization}/-/users/groups?action=open-edit-user-group-dialog&groupName=${name}`,
      "_blank",
    );
  }
</script>

<div
  class={cn("flex items-center gap-2 py-2", {
    "pl-2": leftSpacing,
  })}
>
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
    <span class="text-sm font-medium text-gray-900 flex flex-row gap-x-1">
      {name}
      <span class="text-gray-500 font-normal">
        {isCurrentUser ? "(You)" : ""}
      </span>
      {#if showGuestChip || role === OrgUserRoles.Guest}
        <Chip type="amber" label="Guest" compact readOnly>
          <svelte:fragment slot="body">Guest</svelte:fragment>
        </Chip>
      {/if}
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
      {#if showManage}
        <button
          type="button"
          class="text-xs text-primary-600 font-medium cursor-pointer hover:text-primary-700"
          on:click={handleManageClick}>Manage</button
        >
      {/if}
    </div>
  </div>
</div>
