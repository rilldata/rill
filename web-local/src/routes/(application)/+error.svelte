<script lang="ts">
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";

  $: statusCode = $page.status;
  $: header = ($page.error && $page.error.message) || "An error occurred";
  $: href = $previewModeStore ? "/dashboards" : "/";
</script>

<ErrorPage {statusCode} {header} {href} />
