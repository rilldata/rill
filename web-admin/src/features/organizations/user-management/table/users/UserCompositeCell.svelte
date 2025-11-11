<script lang="ts">
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";

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
