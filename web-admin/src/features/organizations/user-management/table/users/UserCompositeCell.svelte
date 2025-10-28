<script lang="ts">
  import AvatarListItem from "@rilldata/web-admin/features/organizations/user-management/AvatarListItem.svelte";

  export let name: string;
  export let email: string;
  export let isCurrentUser: boolean;
  export let pendingAcceptance: boolean;
  export let photoUrl: string | null;
  export let role: string;
</script>

<AvatarListItem
  {name}
  {email}
  {isCurrentUser}
  {pendingAcceptance}
  {photoUrl}
  {role}
/>
