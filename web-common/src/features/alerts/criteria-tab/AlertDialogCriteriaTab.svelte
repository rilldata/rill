<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import { translateFilter } from "@rilldata/web-common/features/alerts/alert-filter-utils";
  import CriteriaGroup from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaGroup.svelte";
  import DataPreview from "@rilldata/web-common/features/alerts/DataPreview.svelte";

  export let formState: any; // svelte-forms-lib's FormState

  const { form } = formState;
</script>

<div class="flex flex-col gap-y-3">
  <FormSection
    description="Trigger alert when these conditions are met"
    title="Criteria"
  >
    <CriteriaGroup {formState} />
  </FormSection>
  <FormSection title="Alert Preview">
    <DataPreview
      metricsView={$form["metricsViewName"]}
      criteria={translateFilter($form["criteria"], $form["criteriaOperation"])}
      dimension={$form["splitByDimension"]}
      filter={$form["whereFilter"]}
      measure={$form["measure"]}
      timeRange={$form["timeRange"]}
    />
  </FormSection>
</div>
