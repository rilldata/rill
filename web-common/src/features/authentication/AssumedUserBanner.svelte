<script lang="ts">
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { createLocalServiceGetCurrentUser } from "@rilldata/web-common/runtime-client/local-service";

  const user = createLocalServiceGetCurrentUser();
</script>

{#if $user.data?.assumedUserEmail}
  <div class="bg-yellow-100 py-1 w-full">
    <div class="flex flex-row items-center mx-auto w-fit gap-x-2">
      <InfoCircle />
      <span>Warning: Running action as {$user.data?.assumedUserEmail}</span>
    </div>
  </div>
{/if}
