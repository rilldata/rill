<script lang="ts">
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { createErrorPagePropsFromRoutingError } from "../features/errors/error-utils";

  $: errorPageProps = createErrorPagePropsFromRoutingError($page.status);
</script>

<ErrorPage
  statusCode={errorPageProps.statusCode}
  header={errorPageProps.header}
  body={errorPageProps.body}
/>
