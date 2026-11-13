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
  import {
    alertDeleteMutationFactory,
    getAlertMutationFactory,
  } from "@rilldata/web-common/features/alerts/AlertFormMetadataProvider.ts";
  import { createFormMetadataProvider } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";

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

  // The dialog is mounted fresh for each alert, so deriving these once at init is safe.
  // svelte-ignore state_referenced_locally
  const provider = createFormMetadataProvider(
    runtimeClient,
    getAlertMutationFactory(props.mode === "edit"),
    alertDeleteMutationFactory,
  );
  $effect(() => () => provider.cleanup());

  // svelte-ignore state_referenced_locally
  const exploreName =
    props.mode === "create"
      ? props.exploreName
      : getAlertDashboardName(props.alertSpec);

  const validExploreSpec = useExploreValidSpec(runtimeClient, exploreName);
  const metricsViewName = $derived(
    $validExploreSpec.data?.explore?.metricsView ?? "",
  );

  // svelte-ignore state_referenced_locally
  const exploreStateStore =
    props.mode === "create"
      ? useExploreState(props.exploreName)
      : unwrapQueryData(useAlertDashboardState(runtimeClient, props.alertSpec));

  // The creator of an alert created from an embed has a field for their email, which always gets the alert:
  // in an embed, if rill.yaml configures `embeds.open_url`, since only that email gets the open link to the embedder's app;
  // and for alerts with a saved creator email, which the admin service always keeps as a recipient.
  // Outside of an embed, the field is read-only, since only the creator can change it.
  const isEmbedded = EmbedStore.isEmbedded();
  const embedOpenUrlQuery = createRuntimeServiceGetInstance(
    runtimeClient,
    {},
    {
      query: {
        enabled: isEmbedded,
        select: (data) => data?.instance?.embedOpenUrl ?? "",
      },
    },
  );
  // svelte-ignore state_referenced_locally
  const hasSavedEmbedOwnerEmail =
    props.mode === "edit" &&
    !!props.alertSpec.annotations?.admin_embed_owner_email;
  const showEmbedOwnerEmail = $derived(
    hasSavedEmbedOwnerEmail || (isEmbedded && !!$embedOpenUrlQuery.data),
  );

  const initialValues = $derived.by(() => {
    const exploreState = $exploreStateStore;
    if (
      provider.isLoading ||
      (isEmbedded && $embedOpenUrlQuery.isPending) ||
      !exploreState ||
      Object.keys(exploreState).length === 0
    ) {
      return undefined;
    }
    return props.mode === "create"
      ? getNewAlertInitialFormValues(
          metricsViewName,
          exploreName,
          exploreState,
          provider.userEmail,
          showEmbedOwnerEmail,
        )
      : getExistingAlertInitialFormValues(
          props.alertSpec,
          metricsViewName,
          provider.userEmail,
          showEmbedOwnerEmail,
        );
  });
</script>

{#if initialValues}
  <AlertForm
    {props}
    {provider}
    {initialValues}
    {showEmbedOwnerEmail}
    {onClose}
    {onCancel}
  />
{/if}
