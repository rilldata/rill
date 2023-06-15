<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import { ButtonToggleGroup, GroupButton } from "../index.ts";
  import Delta from "../../icons/Delta.svelte";
  import PieChart from "../../icons/PieChart.svelte";

  type ButtonProps = {
    selectionRequired: boolean;
    defaultKey: number | string;
    disabledKeys: (number | string)[];
    bgDark: boolean;
    active: boolean;
  };

  const buttonProps: ButtonProps[] = [];

  for (const selectionRequired of [true, false]) {
    for (const defaultKey of [1, 2, undefined]) {
      for (const disabledKeys of [[1], [2], [1, 2], []]) {
        // for (const disabledKeys of [[1, 2
        //   for (const bgDark of [true, false]) {
        //     for (const active of [true, false]) {
        buttonProps.push({
          selectionRequired,
          defaultKey,
          disabledKeys,
          // bgDark,
          // active,
        });
      }
      //   }
      // }
    }
  }
</script>

<Meta title="Button toggle group stories" />

<Story name="Button toggle group, 2 sub-buttons, no selection required">
  <ButtonToggleGroup>
    <GroupButton key={1}>
      <Delta />%
    </GroupButton>
    <GroupButton key={2}>
      <PieChart />%
    </GroupButton>
  </ButtonToggleGroup>
</Story>

<Story name="Button toggle group, 4 sub-buttons, selection required">
  <ButtonToggleGroup selectionRequired>
    <GroupButton key={1}>
      <Delta />%
    </GroupButton>
    <GroupButton key={2}>
      <PieChart />%
    </GroupButton>
    <GroupButton key={3}>
      <PieChart />%
    </GroupButton>
    <GroupButton key={4}>
      <PieChart />%
    </GroupButton>
  </ButtonToggleGroup>
</Story>

<Template let:args>
  <table>
    <tr>
      <td />
      <td>selectionRequired</td>
      <td>defaultKey</td>
      <td>disabledKeys</td>
      <td>bgDark</td>
      <td>active</td>
    </tr>
    {#each buttonProps as props}
      <tr>
        <td>
          <ButtonToggleGroup {...props}>
            <GroupButton key={1}>
              <Delta />%
            </GroupButton>
            <GroupButton key={2}>
              <PieChart />%
            </GroupButton>
          </ButtonToggleGroup>
        </td>
        <td>{props.selectionRequired}</td>
        <td>{props.defaultKey}</td>
        <td>{props.disabledKeys}</td>
        <td>{props.bgDark}</td>
        <td>{props.active}</td>
      </tr>
    {/each}
  </table>
</Template>

<Story name="Button toggle group variations, two sub-buttons" />

<style>
  td {
    padding: 5px;
  }
  td:first-child {
    padding-right: 40px;
  }
</style>
