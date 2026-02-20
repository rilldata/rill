<script lang="ts">
  import { page } from "$app/stores";
  import APIsTable from "@rilldata/web-admin/features/apis/listing/APIsTable.svelte";
  import { useAPIs } from "@rilldata/web-admin/features/apis/selectors";
  import ProjectPage from "@rilldata/web-admin/features/projects/ProjectPage.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);

  $: ({
    params: { organization, project },
  } = $page);

  $: query = useAPIs(instanceId);

  $: ({ data } = $query);

  // TODO: test this page with a project that has API resources deployed
  $: apis = data?.resources ?? [];
</script>

<ProjectPage {query} kind="API">
  <APIsTable {organization} {project} data={apis} slot="table" />
</ProjectPage>
