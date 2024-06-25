<script lang="ts">
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { createErrorPagePropsFromRoutingError } from "../features/errors/error-utils";

  $: ({ status, error } = $page);

  $: errorPageProps = createErrorPagePropsFromRoutingError(
    status,
    error.message,
  );
</script>

<ErrorPage
  statusCode={errorPageProps.statusCode}
  header={errorPageProps.header}
  body={errorPageProps.body}
  detail={errorPageProps?.detail}
/>
