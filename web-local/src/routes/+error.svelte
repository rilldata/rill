<script lang="ts">
  import { page } from "$app/stores";

  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";

  $: statusCode = $page.status;
  $: header = ($page.error && $page.error.message) || "An error occurred";
  $: body = "";
</script>

<ErrorPage {statusCode} {header} {body} />
