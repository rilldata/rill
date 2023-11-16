<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import { action } from "@storybook/addon-actions";

  import LeaderboardListItem from "../LeaderboardListItem.svelte";

  import type { LeaderboardItemData } from "../leaderboard-utils";
  import { LeaderboardContextColumn } from "../../leaderboard-context-column";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

  const atLeastOneActive = true;
  const filterExcludeMode = true;
  const isSummableMeasure = true;
  const referenceValue = 400;
  const formatPreset = "humanize";

  const itemData: LeaderboardItemData = {
    dimensionValue: "Widget Co.",
    value: 300,
    pctOfTotal: 0.4,
    prevValue: 200,
    deltaRel: 0.5,
    deltaAbs: 100,
    selectedIndex: -1,
  };
  const defaultArgs = {
    itemData,
    contextColumn: LeaderboardContextColumn.HIDDEN,
    atLeastOneActive: false,
    isBeingCompared: false,
    filterExcludeMode,
    formatPreset,
    isSummableMeasure,
    referenceValue,
  };
</script>

<Meta
  title="Leaderboard/LeaderboardListItem"
  argTypes={{
    contextColumn: {
      options: Object.values(LeaderboardContextColumn),
      control: {
        type: "inline-radio",
        labels: LeaderboardContextColumn,
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
    formatPreset: {
      control: {
        type: "inline-radio",
        labels: FormatPreset,
      },
      options: Object.values(FormatPreset),
    },
  }}
/>

<Template let:args>
  <div style:width="365px">
    <LeaderboardListItem
      itemData={{
        label: args.label,
        value: args.value,
        selected: args.selected,
        comparisonValue: args.comparisonValue,
      }}
      {atLeastOneActive}
      {filterExcludeMode}
      {isSummableMeasure}
      {referenceValue}
      {formatPreset}
      on:select-item={(evt) => {
        action("select-item")(evt.detail);
      }}
      {...args}
    />
  </div>
</Template>

<Story name="LeaderboardListItem (with controls)" args={defaultArgs} />
