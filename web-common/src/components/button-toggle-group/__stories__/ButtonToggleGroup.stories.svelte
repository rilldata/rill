<script lang="ts">
  import { Meta, Story } from "@storybook/addon-svelte-csf";

  import { ButtonToggleGroup, SubButton } from "../index";
  import Delta from "../../icons/Delta.svelte";
  import PieChart from "../../icons/PieChart.svelte";

  const pieTooltips = {
    selected: "Hide percent of total",
    unselected: "Show percent of total",
    disabled: "To show percent of total, show top values by a summable metric",
  };

  const deltaTooltips = {
    selected: "Hide percent change",
    unselected: "Show percent change",
    disabled: "To show percent change, select a comparison period abovec",
  };
</script>

<Meta title="Button toggle group stories" />

<Story name="Button toggle group variations (tables)">
  {#each [[], [1], [2], [1, 2]] as disabledKeys}
    disabledKeys: {JSON.stringify(disabledKeys)}
    <table>
      <tr>
        <td />
        <td>selectionRequired:<br />true</td>
        <td>selectionRequired:<br />false</td>
      </tr>
      {#each [1, 2, undefined] as defaultKey}
        <tr>
          <td>defaultKey: {defaultKey}</td>
          {#each [true, false] as selectionRequired}
            <td>
              <ButtonToggleGroup
                {...{ disabledKeys, defaultKey, selectionRequired }}
              >
                <SubButton key={1} tootips={deltaTooltips}>
                  <Delta />%
                </SubButton>
                <SubButton key={2} tootips={pieTooltips}>
                  <PieChart />%
                </SubButton>
              </ButtonToggleGroup>
            </td>
          {/each}
        </tr>
      {/each}
    </table>
    <br /> <br /> <br />
  {/each}
</Story>

<Story name="Button toggle group, 2 sub-buttons, no selection required">
  <ButtonToggleGroup>
    <SubButton key={1} tootips={deltaTooltips}>
      <Delta />%
    </SubButton>
    <SubButton key={2} tootips={pieTooltips}>
      <PieChart />%
    </SubButton>
  </ButtonToggleGroup>
</Story>

<Story name="Button toggle group, 4 sub-buttons, selection required">
  <ButtonToggleGroup selectionRequired>
    <SubButton key={1}>
      <Delta />%
    </SubButton>
    <SubButton key={2}>
      <PieChart />%
    </SubButton>
    <SubButton key={3}>
      <PieChart />%
    </SubButton>
    <SubButton key={4}>
      <PieChart />%
    </SubButton>
  </ButtonToggleGroup>
</Story>

<style>
  td {
    padding: 5px;
  }
  td:first-child {
    padding-right: 40px;
  }
</style>
