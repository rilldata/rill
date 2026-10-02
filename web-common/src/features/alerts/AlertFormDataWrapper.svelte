<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    getAlertDashboardName,
    unwrapQueryData,
    useAlertDashboardState,
  } from "@rilldata/web-admin/features/alerts/selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import { getNewAlertInitialFormValues } from "@rilldata/web-common/features/alerts/create-alert-utils.ts";
  import { getExistingAlertInitialFormValues } from "@rilldata/web-common/features/alerts/extract-alert-form-values.ts";
  import AlertForm, {
    type CreateAlertProps,
    type EditAlertProps,
  } from "@rilldata/web-common/features/alerts/AlertForm.svelte";
  import { createFormMetadataProvider } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
  import { getAlertMutationFactory } from "@rilldata/web-common/features/alerts/AlertFormMetadataProvider.ts";

  let {
    onClose,
    onCancel,
    props,
  }: {
    onClose: () => void;
    onCancel: () => void;
    props: CreateAlertProps | EditAlertProps;
  } = $props();

  const runtimeClient = useRuntimeClient();

  // The wrapper is mounted fresh for each alert, so deriving these once at init is safe.
  // svelte-ignore state_referenced_locally
  const provider = createFormMetadataProvider(
    runtimeClient,
    getAlertMutationFactory(props.mode === "edit"),
  );
  $effect(() => () => provider.cleanup());

  const exploreName = $derived(
    props.mode === "create"
      ? props.exploreName
      : getAlertDashboardName(props.alertSpec),
  );

  const validExploreSpec = $derived(
    useExploreValidSpec(runtimeClient, exploreName),
  );
  const metricsViewName = $derived(
    $validExploreSpec.data?.explore?.metricsView ?? "",
  );

  // svelte-ignore state_referenced_locally
  const exploreStateStore =
    props.mode === "create"
      ? useExploreState(props.exploreName)
      : unwrapQueryData(useAlertDashboardState(runtimeClient, props.alertSpec));

  const initialValues = $derived.by(() => {
    const exploreState = $exploreStateStore;
    if (
      provider.isLoading ||
      !exploreState ||
      Object.keys(exploreState).length === 0
    )
      return undefined;

    return props.mode === "create"
      ? getNewAlertInitialFormValues(
          metricsViewName,
          exploreName,
          exploreState,
          provider.userEmail,
        )
      : getExistingAlertInitialFormValues(props.alertSpec, metricsViewName);
  });
</script>

{#if initialValues}
  <AlertForm {props} {provider} {initialValues} {onClose} {onCancel} />
{/if}
