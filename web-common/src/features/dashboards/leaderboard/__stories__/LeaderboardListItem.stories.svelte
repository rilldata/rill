<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import { action } from "@storybook/addon-actions";

  import LeaderboardListItem from "../LeaderboardListItem.svelte";

  import { NicelyFormattedTypes } from "../../humanize-numbers";

  const atLeastOneActive = true;
  const filterExcludeMode = true;
  const isSummableMeasure = true;
  const referenceValue = 400;
  const formatPreset = "humanize";

  const defaultArgs = {
    label: "item label",
    value: 300,
    selected: true,
    comparisonValue: 200,
    showContext: "time",
    filterExcludeMode: false,
    isSummableMeasure: true,
    referenceValue: 400,
    unfilteredTotal: 1000,
    formatPreset: "humanize",
  };
</script>

<Meta
  title="Leaderboard/LeaderboardListItem"
  argTypes={{
    label: { control: "text" },
    value: 300,
    selected: {
      control: {
        type: "boolean",
      },
    },
    comparisonValue: 200,
    showContext: {
      control: {
        type: "inline-radio",
      },
      options: ["time", "percent", "false"],
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
      },
      options: [
        NicelyFormattedTypes.HUMANIZE,
        NicelyFormattedTypes.PERCENTAGE,
        NicelyFormattedTypes.CURRENCY,
        NicelyFormattedTypes.NONE,
      ],
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
