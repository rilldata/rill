<script lang="ts">
  import APIsTable from "@rilldata/web-admin/features/apis/listing/APIsTable.svelte";
  import { useAPIs } from "@rilldata/web-admin/features/apis/selectors";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  $: query = useAPIs(instanceId);

  $: ({ data } = $query);

  $: apis = data?.resources ?? [];
</script>

<ProjectPage {query} kind="API">
  <APIsTable data={apis} slot="table" />
</ProjectPage>
