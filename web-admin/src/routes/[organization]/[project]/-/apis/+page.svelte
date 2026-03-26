<script lang="ts">
  import APIsTable from "@rilldata/web-admin/features/apis/listing/APIsTable.svelte";
  import { useAPIs } from "@rilldata/web-admin/features/apis/selectors";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();

  $: query = useAPIs(runtimeClient);

  $: ({ data } = $query);

  $: apis = data?.resources ?? [];
</script>

<ProjectPage {query} kind="API">
  <APIsTable data={apis} slot="table" />
</ProjectPage>
