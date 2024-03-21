<script lang="ts">
  import { runtime } from "../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../entity-management/actions";
  import { EntityType } from "../entity-management/types";
  import { useCustomDashboardFileNames } from "./selectors";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

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

<DropdownMenu.Item on:click={handleDeleteCustomDashboard}>
  Delete custom dashboard
</DropdownMenu.Item>
