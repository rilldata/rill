<script lang="ts">
  import { WorkspaceContainer } from "../../layout/workspace";

  interface OnboardingStep {
    heading: string;
    description: string;
  }

  const steps: OnboardingStep[] = [
    {
      heading: "Import your data source",
      description:
        "Click 'Add data' or drag a file (Parquet, NDJSON, or CSV) into this window.",
    },
    {
      heading: "Model your sources into one big table",
      description:
        "Build intuition about your sources and use SQL to model them into an analytics-ready resource.",
    },
    {
      heading: "Define your metrics and dimensions",
      description:
        "Define aggregate metrics and break out dimensions for your modeled data.",
    },
    {
      heading: "Explore your metrics dashboard",
      description:
        "Interactively explore line charts and leaderboards to uncover insights.",
    },
  ];
</script>

<WorkspaceContainer top="0px" assetID="onboarding" inspector={false}>
  <div class="pt-20 px-8 flex flex-col gap-y-6 items-center" slot="body">
    <div class="text-center">
      <div class="font-bold">Getting started</div>
      <p>Building data intuition at every step of analysis</p>
    </div>
    <ol
      class="w-full lg:w-2/3 xl:w-1/2 2xl:w-2/5 flex flex-col gap-y-4 px-9 pt-9 pb-[60px] bg-gray-50 rounded-lg border border-gray-200"
    >
      {#each steps as step, i (step.heading)}
        <li class="flex gap-x-0.5">
          <span class="font-bold">{i + 1}.</span>
          <div class="flex flex-col items-start gap-y-0.5">
            <h5 class="font-bold">{step.heading}</h5>
            <p>{step.description}</p>
          </div>
        </li>
      {/each}
    </ol>
  </div>
</WorkspaceContainer>
