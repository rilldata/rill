<script lang="ts">
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    getAlertDashboardName,
    unwrapQueryData,
    useAlertDashboardState,
  } from "@rilldata/web-admin/features/alerts/selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import { isAIAlertSpec } from "@rilldata/web-common/features/alerts/ai-alert-utils.ts";
  import { getNewAlertInitialFormValues } from "@rilldata/web-common/features/alerts/create-alert-utils.ts";
  import { getExistingAlertInitialFormValues } from "@rilldata/web-common/features/alerts/extract-alert-form-values.ts";
  import AlertForm, {
    type CreateAlertProps,
    type EditAlertProps,
  } from "@rilldata/web-common/features/alerts/AlertForm.svelte";
  import { derived } from "svelte/store";

  export let onClose: () => void;
  export let onCancel: () => void;
  export let props: CreateAlertProps | EditAlertProps;

  const user = createAdminServiceGetCurrentUser();
  const runtimeClient = useRuntimeClient();

  // The dialog opens for one alert or dashboard, so these are read once.
  const exploreName =
    props.mode === "create"
      ? props.exploreName
      : getAlertDashboardName(props.alertSpec);
  const validExploreSpec = useExploreValidSpec(runtimeClient, exploreName);

  const exploreStateStore =
    props.mode === "create"
      ? useExploreState(props.exploreName)
      : unwrapQueryData(useAlertDashboardState(runtimeClient, props.alertSpec));

  // An AI alert created through the API has no captured dashboard state, and editing it does not need one.
  const needsExploreState =
    props.mode === "create" || !isAIAlertSpec(props.alertSpec);

  // The form captures its initial values once, so they wait for everything they are built from, the explore's metrics view included:
  // without it the alert would be saved with an empty metrics view, which its managed YAML refs.
  const initialValuesStore = derived(
    [exploreStateStore, user, validExploreSpec],
    ([exploreState, userResp, exploreResp]) => {
      const metricsViewName = exploreResp.data?.explore?.metricsView ?? "";
      if (
        userResp.isPending ||
        !metricsViewName ||
        (needsExploreState &&
          (!exploreState || Object.keys(exploreState).length === 0))
      )
        return {
          isLoading: true,
          data: undefined,
        };

      const initialValues =
        props.mode === "create"
          ? getNewAlertInitialFormValues(
              metricsViewName,
              exploreName,
              exploreState ?? {}, // Defined: creating needs the explore state (see needsExploreState)
              userResp.data?.user,
            )
          : getExistingAlertInitialFormValues(props.alertSpec, metricsViewName);

      return {
        isLoading: false,
        data: initialValues,
      };
    },
  );
  $: ({ data: initialValues, isLoading } = $initialValuesStore);
</script>

{#if !isLoading && initialValues}
  <AlertForm {props} {initialValues} {onClose} {onCancel} />
{/if}
