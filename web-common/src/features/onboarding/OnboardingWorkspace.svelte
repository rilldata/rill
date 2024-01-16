<script lang="ts">
  import {_} from "svelte-i18n";
  import AddSourceModal from "@rilldata/web-common/features/sources/modal/AddSourceModal.svelte";
  import { IconSpaceFixer } from "../../components/button";
  import Button from "../../components/button/Button.svelte";
  import Add from "../../components/icons/Add.svelte";
  import { WorkspaceContainer } from "../../layout/workspace";

  interface OnboardingStep {
    id: string;
    heading: string;
    description: string;
  }

  const steps: OnboardingStep[] = [
    {
      id: "source",
      heading: $_('import-your-data-source'),
      description:
        $_('click-add-data-or-drag-a-file-parquet-ndjson-or-csv-into-this-window'),
    },
    {
      id: "model",
      heading: $_('model-your-sources-into-one-big-table'),
      description:
        $_('build-intuition-about-your-sources-and-use-sql-to-model-them-into-an-analytics-ready-resource'),
    },
    {
      id: "metrics",
      heading: $_('define-your-metrics-and-dimensions'),
      description:
        $_('define-aggregate-metrics-and-break-out-dimensions-for-your-modeled-data'),
    },
    {
      id: "dashboard",
      heading: $_('explore-your-metrics-dashboard'),
      description:
        $_('interactively-explore-line-charts-and-leaderboards-to-uncover-insights'),
    },
  ];

  let showAddSourceModal = false;
  function openAddSourceModal() {
    showAddSourceModal = true;
  }
</script>

<WorkspaceContainer top="0px" assetID="onboarding" inspector={false}>
  <div class="pt-20 px-8 flex flex-col gap-y-6 items-center" slot="body">
    <div class="text-center">
      <div class="font-bold">{$_('getting-started')}</div>
      <p>{$_('building-data-intuition-at-every-step-of-analysis')}</p>
    </div>
    <ol
      class="max-w-fit flex flex-col gap-y-4 px-9 pt-9 pb-[60px] bg-gray-50 rounded-lg border border-gray-200"
    >
      {#each steps as step, i (step.heading)}
        <li class="flex gap-x-0.5">
          <span class="font-bold">{i + 1}.</span>
          <div class="flex flex-col items-start gap-y-2">
            <div class="flex flex-col items-start gap-y-0.5">
              <h5 class="font-bold">{step.heading}</h5>
              <p>{step.description}</p>
            </div>
            {#if step.id === "source"}
              <Button type="secondary" on:click={openAddSourceModal}>
                <IconSpaceFixer pullLeft><Add /></IconSpaceFixer>
                <span>{$_('add-data')}</span>
              </Button>
            {/if}
          </div>
        </li>
      {/each}
    </ol>
    <AddSourceModal
      open={showAddSourceModal}
      on:close={() => (showAddSourceModal = false)}
    />
  </div>
</WorkspaceContainer>
