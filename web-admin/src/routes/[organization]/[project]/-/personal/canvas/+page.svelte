<script lang="ts">
  import { page } from "$app/state";
  import VirtualFileEditor from "@rilldata/web-admin/features/virtual-file-editor/VirtualFileEditor.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

  let org = $derived(page.params.organization);
  let project = $derived(page.params.project);
</script>

<VirtualFileEditor {org} {project} name="canvas" kind={ResourceKind.Canvas} />
