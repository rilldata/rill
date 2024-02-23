<script lang="ts">
  import { translateFilter } from "@rilldata/web-common/features/alerts/alert-filter-utils";
  import { AlertIntervalOptions } from "@rilldata/web-common/features/alerts/data-tab/intervals";
  import FormSection from "../../../components/forms/FormSection.svelte";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useMetricsView } from "../../dashboards/selectors";
  import AlertDataPreview from "web-common/src/features/alerts/AlertDataPreview.svelte";
  import NoFiltersSelected from "./NoFiltersSelected.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form, errors, handleChange } = formState;

  $: metricsView = useMetricsView(
    $runtime.instanceId,
    $form["metricsViewName"],
  );

  $: measureOptions =
    $metricsView.data?.measures?.map((m) => ({
      value: m.name as string,
      label: m.label?.length ? m.label : m.expression,
    })) ?? [];
  $: dimensionOptions =
    $metricsView.data?.dimensions?.map((d) => ({
      value: d.name as string,
      label: d.label?.length ? d.label : d.expression,
    })) ?? [];

  $: hasAtLeastOneFilter = $form.whereFilter.cond.exprs.length > 0;
</script>

<div class="flex flex-col gap-y-3">
  <FormSection title="Alert name">
    <InputV2
      error={$errors["name"]}
      id="name"
      on:change={handleChange}
      placeholder="My alert"
      value={$form["name"]}
    />
  </FormSection>
  <FormSection
    description={hasAtLeastOneFilter
      ? "These are inherited from the underlying dashboard view."
      : ""}
    title="Filters"
  >
    {#if hasAtLeastOneFilter}
      TODO: read-only FilterChips without StateManagers ctx
      <!-- <FilterChips readOnly /> -->
    {:else}
      <NoFiltersSelected />
    {/if}
  </FormSection>
  <FormSection
    description="Select the measures you want to monitor."
    title="Alert data"
  >
    <Select
      bind:value={$form["measure"]}
      id="measure"
      label="Measure"
      options={measureOptions}
      placeholder="Select a measure"
    />
    <Select
      bind:value={$form["splitByDimension"]}
      id="splitByDimension"
      label="Split by dimension"
      optional
      options={dimensionOptions}
      placeholder="Select a dimension"
    />
    <Select
      bind:value={$form["splitByTimeGrain"]}
      id="splitByTimeGrain"
      label="Split by time grain"
      optional
      options={AlertIntervalOptions}
      placeholder="Select a time grain"
    />
  </FormSection>
  <FormSection title="Data preview">
    <AlertDataPreview
      criteria={translateFilter($form["criteria"], $form["criteriaOperation"])}
      measure={$form["measure"]}
      metricsViewName={$form["metricsViewName"]}
      splitByDimension={$form["splitByDimension"]}
      splitByTimeGrain={$form["splitByTimeGrain"]}
      timeRange={$form["timeRange"]}
      whereFilter={$form["whereFilter"]}
    />
  </FormSection>
</div>
