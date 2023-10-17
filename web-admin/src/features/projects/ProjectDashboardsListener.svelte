<script lang="ts">
  import { listenAndInvalidateDashboards } from "@rilldata/web-admin/features/projects/dashboards";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import type { Unsubscriber } from "svelte/store";

  const queryClient = useQueryClient();

  let unsubscribe: Unsubscriber;
  $: if ($runtime?.instanceId) {
    unsubscribe?.();
    unsubscribe = listenAndInvalidateDashboards(
      queryClient,
      $runtime?.instanceId
    );
  }
</script>

<slot />
