<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import { action } from "@storybook/addon-actions";
  import LeaderboardHeader from "../LeaderboardHeader.svelte";
  // import { FormatPreset } from "../../humanize-numbers";

  // const atLeastOneActive = true;
  // const filterExcludeMode = true;
  // const isSummableMeasure = true;
  // const referenceValue = 400;
  // const formatPreset = "humanize";

  const defaultArgs = {
    displayName: "Leaderboard Name",
    dimensionDescription: "dimension description",
    isFetching: false,
    hovered: false,
    showTimeComparison: false,
    showPercentOfTotal: false,
    filterExcludeMode: false,
    sortDescending: true,
    // referenceValue: 400,
    // unfilteredTotal: 1000,
    // formatPreset: "humanize",
  };
</script>

<Meta
  title="Leaderboard/LeaderboardHeader"
  argTypes={{
    displayName: { control: "text" },
    dimensionDescription: { control: "text" },
    isFetching: {
      control: {
        type: "boolean",
      },
    },

    filterExcludeMode: {
      control: {
        type: "boolean",
      },
    },
    isSummableMeasure: {
      control: {
        type: "boolean",
      },
    },
    sortDescending: {
      control: {
        type: "boolean",
      },
    },
  }}
/>

<Template let:args>
  <div style:width="365px">
    <LeaderboardHeader
      on:select-item={(evt) => {
        action("select-item")(evt.detail);
      }}
      {...args}
    />
  </div>
</Template>

<Story name="LeaderboardHeader" args={defaultArgs} />
