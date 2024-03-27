<script lang="ts">
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { runtime } from "../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../entity-management/actions";
  import { EntityType } from "../entity-management/types";
  import { useCustomDashboardRoutes } from "./selectors";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  export let customDashboardName: string;

  $: customDashboardRoutes = useCustomDashboardRoutes($runtime.instanceId);

  async function handleDeleteCustomDashboard() {
    await deleteFileArtifact(
      $runtime.instanceId,
      getFileAPIPathFromNameAndType(customDashboardName, EntityType.Dashboard),
      EntityType.Dashboard,
      $customDashboardRoutes.data ?? [],
    );
  }
</script>

<DropdownMenu.Item on:click={handleDeleteCustomDashboard}>
  Delete custom dashboard
</DropdownMenu.Item>
