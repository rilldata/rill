<script lang="ts">
  import { page } from "$app/stores";
  import EditSession from "@rilldata/web-admin/features/edit-session/EditSession.svelte";

  $: ({ organization, project } = $page.params);
</script>

<div class="h-full">
  <EditSession {organization} {project} />
</div>
