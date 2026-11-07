<script lang="ts">
  import { getAdminServiceListPersonalFilesQueryKey } from "@rilldata/web-admin/client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { onMount } from "svelte";

  let { organization, project }: { organization: string; project: string } =
    $props();

  // A personal file saved outside the page that lists them (e.g. by the AI
  // from the chat) refreshes the list. It is refetched even with nobody
  // watching it: queries don't refetch on mount, so a list opened later would
  // otherwise show its stale copy.
  onMount(() =>
    eventBus.on("personal-files-changed", () => {
      void queryClient.invalidateQueries({
        queryKey: getAdminServiceListPersonalFilesQueryKey(
          organization,
          project,
        ),
        refetchType: "all",
      });
    }),
  );
</script>
