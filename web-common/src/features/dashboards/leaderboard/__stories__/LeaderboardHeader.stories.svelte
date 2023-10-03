<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import { action } from "@storybook/addon-actions";
  import LeaderboardHeader from "../LeaderboardHeader.svelte";
  import { LeaderboardContextColumn } from "../../leaderboard-context-column";
  import { SortType } from "../../proto-state/derived-types";

  const defaultArgs = {
    displayName: "Leaderboard Name",
    dimensionDescription: "dimension description",
    isFetching: false,
    hovered: false,
    // showTimeComparison: false,
    // showPercentOfTotal: false,
    // filterExcludeMode: false,
    // sortDescending: true,

    contextColumn: LeaderboardContextColumn.HIDDEN,
    sortAscending: false,
    sortType: SortType.VALUE,
    isBeingCompared: false,
  };

  console.log("LeaderboardContextColumn", LeaderboardContextColumn);
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

    // filterExcludeMode: {
    //   control: {
    //     type: "boolean",
    //   },
    // },
    // isSummableMeasure: {
    //   control: {
    //     type: "boolean",
    //   },
    // },
    // sortDescending: {
    //   control: {
    //     type: "boolean",
    //   },
    // },

    hovered: {
      control: {
        type: "boolean",
      },
    },
    contextColumn: {
      options: Object.values(LeaderboardContextColumn),
      control: {
        type: "select",
        labels: LeaderboardContextColumn,
      },
    },
    sortAscending: {
      control: {
        type: "boolean",
      },
    },
    sortType: {
      options: Object.values(SortType),
      control: {
        type: "select",
        labels: SortType,
      },
    },
    isBeingCompared: {
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
