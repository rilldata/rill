<script lang="ts">
  import { MenuItem } from "@rilldata/web-common/components/menu";

  export let customDashboardName: string;

  function handleDeleteCustomDashboard() {
    console.log("delete custom dashboard", customDashboardName);
  }
</script>

<MenuItem icon on:select={handleDeleteCustomDashboard} propogateSelect={false}>
  Delete custom dashboard
</MenuItem>
