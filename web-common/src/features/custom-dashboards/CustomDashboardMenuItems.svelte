<script lang="ts">
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { runtime } from "../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../entity-management/actions";
  import { EntityType } from "../entity-management/types";
  import { useCustomDashboardRoutes } from "./selectors";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { goto } from "$app/navigation";
  import { getNextRoute } from "../models/utils/navigate-to-next";

  export let customDashboardName: string;
  export let open: boolean;

  $: customDashboardRoutesQuery = useCustomDashboardRoutes($runtime.instanceId);
  $: customDashboardRoutes = $customDashboardRoutesQuery.data ?? [];

  async function handleDeleteCustomDashboard() {
    try {
      await deleteFileArtifact(
        $runtime.instanceId,
        getFileAPIPathFromNameAndType(
          customDashboardName,
          EntityType.Dashboard,
        ),
        EntityType.Dashboard,
      );

      if (open) await goto(getNextRoute(customDashboardRoutes));
    } catch (error) {
      console.error(error);
    }
  }
</script>

<DropdownMenu.Item on:click={handleDeleteCustomDashboard}>
  Delete custom dashboard
</DropdownMenu.Item>
