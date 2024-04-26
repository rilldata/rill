<script lang="ts">
  import { IconSpaceFixer } from "../../components/button";
  import Button from "../../components/button/Button.svelte";
  import Add from "../../components/icons/Add.svelte";
  import { createRuntimeServiceGetInstance } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { addSourceModal } from "../sources/modal/add-source-visibility";

  let steps: OnboardingStep[];
  $: instance = createRuntimeServiceGetInstance($runtime.instanceId);
  $: olapConnector = $instance.data?.instance?.olapConnector;
  $: if (olapConnector) {
    steps = olapConnector === "duckdb" ? duckDbSteps : nonDuckDbSteps;
  }

  interface OnboardingStep {
    id: string;
    heading: string;
    description: string;
  }

  // Onboarding steps for DuckDB OLAP driver
  const duckDbSteps: OnboardingStep[] = [
    {
      id: "source",
      heading: "Import your data source",
      description:
        "Click 'Add data' or drag a file (Parquet, NDJSON, or CSV) into this window.",
    },
    {
      id: "model",
      heading: "Model your sources into one big table",
      description:
        "Build intuition about your sources and use SQL to model them into an analytics-ready resource.",
    },
    {
      id: "metrics",
      heading: "Define your metrics and dimensions",
      description:
        "Define aggregate metrics and break out dimensions for your modeled data.",
    },
    {
      id: "dashboard",
      heading: "Explore your metrics dashboard",
      description:
        "Interactively explore line charts and leaderboards to uncover insights.",
    },
  ];

  // Onboarding steps for non-DuckDB OLAP drivers (ClickHouse, Druid)
  const nonDuckDbSteps: OnboardingStep[] = [
    {
      id: "table",
      heading: "Explore your tables",
      description:
        "Find your database tables in the left-hand-side navigational sidebar.",
    },
    {
      id: "metrics",
      heading: "Define your metrics and dimensions",
      description:
        "Define aggregate metrics and break out dimensions for your tables.",
    },
    {
      id: "dashboard",
      heading: "Explore your metrics dashboard",
      description:
        "Interactively explore line charts and leaderboards to uncover insights.",
    },
  ];
</script>

<div
  class="pt-20 px-8 flex flex-col gap-y-6 items-center bg-gray-100 size-full"
>
  <div class="text-center">
    <div class="font-bold">Getting started</div>
    <p>Building data intuition at every step of analysis</p>
  </div>
  <ol
    class="max-w-fit flex flex-col gap-y-4 px-9 pt-9 pb-[60px] bg-gray-50 rounded-lg border border-gray-200"
  >
    {#if olapConnector}
      {#each steps as step, i (step.heading)}
        <li class="flex gap-x-0.5">
          <span class="font-bold">{i + 1}.</span>
          <div class="flex flex-col items-start gap-y-2">
            <div class="flex flex-col items-start gap-y-0.5">
              <h5 class="font-bold">{step.heading}</h5>
              <p>{step.description}</p>
            </div>
            {#if step.id === "source"}
              <Button type="secondary" on:click={addSourceModal.open}>
                <IconSpaceFixer pullLeft><Add /></IconSpaceFixer>
                <span>Add data</span>
              </Button>
            {/if}
          </div>
        </li>
      {/each}
    {/if}
  </ol>
</div>
