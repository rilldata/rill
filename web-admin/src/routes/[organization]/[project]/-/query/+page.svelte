<script lang="ts">
  import QueryWorkspace from "@rilldata/web-common/features/query/QueryWorkspace.svelte";
  import { page } from "$app/stores";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="size-full overflow-hidden">
  <QueryWorkspace projectId="{organization}/{project}" />
</div>
