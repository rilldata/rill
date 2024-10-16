<script lang="ts">
  import AvatarCircleList from "./AvatarCircleList.svelte";

  export let name: string;
  export let email: string;
  export let isCurrentUser: boolean;
  export let pendingAcceptance: boolean;
  export let photoUrl: string | null;
</script>

<AvatarCircleList
  {name}
  {email}
  {isCurrentUser}
  {pendingAcceptance}
  {photoUrl}
/>
