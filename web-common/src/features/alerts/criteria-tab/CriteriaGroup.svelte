<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CriteriaForm from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaForm.svelte";
  import { CriteriaGroupOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import { getEmptyMeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
  import { Trash2Icon } from "lucide-svelte";
  import type { SuperForm } from "sveltekit-superforms/client";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let timeControls: TimeControls;

  $: ({ form } = superFormInstance);

  function handleDeleteCriteria(index: number) {
    $form["criteria"] = $form["criteria"].filter((_, i: number) => i !== index);
  }

  function handleAddCriteria() {
    $form["criteria"] = $form["criteria"].concat({
      ...getEmptyMeasureFilterEntry(),
      measure: $form.measure,
    });
  }
</script>

{#if $form["criteria"]}
  <div class="flex flex-col gap-2">
    {#each $form["criteria"] as _, index (index)}
      {#if index > 0}
        <div class="flex flex-row items-center justify-center">
          <div class="mr-2">
            <Select
              bind:value={$form["criteriaOperation"]}
              id="field"
              label=""
              ariaLabel="Criteria group operation"
              options={CriteriaGroupOptions}
              placeholder="Measure"
            />
          </div>
        </div>
      {/if}
      <div class="flex flex-col gap-2">
        <div class="flex flex-row items-center">
          <div class="w-full text-lg">{index + 1}</div>
          <button class="mr-2" on:click={() => handleDeleteCriteria(index)}>
            <Trash2Icon size="16px" />
          </button>
        </div>
        <CriteriaForm {superFormInstance} {timeControls} {index} />
      </div>
    {/each}
    <Button type="outlined" onClick={handleAddCriteria}>+ Add Criteria</Button>
  </div>
{/if}
