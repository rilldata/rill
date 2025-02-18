<script lang="ts">
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createLocalServiceGetCurrentUser } from "@rilldata/web-common/runtime-client/local-service";

  const { disableCloud } = featureFlags;

  const user = createLocalServiceGetCurrentUser({
    query: {
      enabled: !$disableCloud,
    },
  });
</script>

{#if $user.data?.isRepresentingUser}
  <div class="bg-yellow-100 py-1 w-full">
    <div class="flex flex-row items-center mx-auto w-fit gap-x-2">
      <InfoCircle />
      <span>Warning: Running action as {$user.data?.user?.email}</span>
    </div>
  </div>
{/if}
