<script lang="ts">
  import { listenAndInvalidateDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy } from "svelte";
  import type { Unsubscriber } from "svelte/store";

  const queryClient = useQueryClient();

  let unsubscribe: Unsubscriber;
  $: if ($runtime?.instanceId) {
    unsubscribe?.();
    unsubscribe = listenAndInvalidateDashboards(
      queryClient,
      $runtime?.instanceId,
    );
  }

  onDestroy(() => {
    unsubscribe?.();
  });
</script>

<slot />
