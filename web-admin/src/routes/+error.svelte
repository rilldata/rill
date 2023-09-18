<script lang="ts">
  import { page } from "$app/stores";
  import { createErrorPagePropsFromRoutingError } from "../features/errors/error-utils";
  import ErrorPage from "../features/errors/ErrorPage.svelte";

  $: errorPageProps = createErrorPagePropsFromRoutingError($page.status);
</script>

<ErrorPage
  statusCode={errorPageProps.statusCode}
  header={errorPageProps.header}
  body={errorPageProps.body}
/>
