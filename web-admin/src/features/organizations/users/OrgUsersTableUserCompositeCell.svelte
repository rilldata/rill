<script lang="ts">
  import AvatarListItem from "./AvatarListItem.svelte";

  export let name: string;
  export let email: string;
  export let isCurrentUser: boolean;
  export let pendingAcceptance: boolean;
  export let photoUrl: string | null;
</script>

<AvatarListItem {name} {email} {isCurrentUser} {pendingAcceptance} {photoUrl} />
