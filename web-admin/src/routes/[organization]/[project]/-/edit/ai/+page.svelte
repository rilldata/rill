<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import AiChatPage from "@rilldata/web-common/features/preview-mode/AiChatPage.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: branch = $page.url.pathname.match(/\/@([^/]+)/)?.[1];
  $: basePath = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit/ai`;
</script>

<AiChatPage {basePath} />
