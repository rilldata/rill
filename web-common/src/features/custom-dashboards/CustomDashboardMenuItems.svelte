<script lang="ts">
  import { MenuItem } from "@rilldata/web-common/components/menu";
  import { runtime } from "../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../entity-management/actions";
  import { EntityType } from "../entity-management/types";
  import { useCustomDashboardFileNames } from "./selectors";

  export let customDashboardName: string;

  $: customDashboardFileNames = useCustomDashboardFileNames(
    $runtime.instanceId,
  );

  async function handleDeleteCustomDashboard() {
    await deleteFileArtifact(
      $runtime.instanceId,
      customDashboardName,
      EntityType.Dashboard,
      $customDashboardFileNames.data ?? [],
    );
  }
</script>

<MenuItem icon on:select={handleDeleteCustomDashboard} propogateSelect={false}>
  Delete custom dashboard
</MenuItem>
