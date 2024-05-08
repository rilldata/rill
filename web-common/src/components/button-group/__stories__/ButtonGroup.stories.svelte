<script lang="ts">
  import { Meta, Story } from "@storybook/addon-svelte-csf";

  import { action } from "@storybook/addon-actions";

  import Delta from "../../icons/Delta.svelte";
  import PieChart from "../../icons/PieChart.svelte";
  import { ButtonGroup, SubButton } from "../index";

  const pieTooltips = {
    selected: "Hide percent of total",
    unselected: "Show percent of total",
    disabled: "To show percent of total, show top values by a summable metric",
  };

  const deltaTooltips = {
    selected: "Hide percent change",
    unselected: "Show percent change",
    disabled: "To show percent change, select a comparison period above",
  };

  const deltaPieCombos = [[], ["delta"], ["pie"], ["delta", "pie"]];
</script>

<Meta
  title="Button group stories"
  argTypes={{
    clickAction: { action: "subbutton-click" },
  }}
/>

<Story name="Button group, selected vs disabled (tabular)">
  <p>
    This story shows the different combinations of selected and disabled props.
    Note that this component does not maintain internal state, so it is up to
    the parent component to manage the state of each subbutton by passing in the
    selected and disabled props, which is why the buttons don't maintain their
    toggle state when you click on them.
  </p>
  <br />
  <p>
    The ButtonGroup component will dispatch a "subbutton-click" event with the
    value of the subbutton that was clicked. You can see this in the action
    logger below. If the subbutton is disabled, the event will not be
    dispatched.
  </p>
  <table>
    <tr>
      <td />
      {#each deltaPieCombos as disabled}
        <td>disabled:<br /> {JSON.stringify(disabled)}</td>
      {/each}
    </tr>
    {#each deltaPieCombos as selected}
      <tr>
        <td>selected: {JSON.stringify(selected)}</td>
        {#each deltaPieCombos as disabled}
          <td>
            <ButtonGroup
              {...{ selected, disabled }}
              on:subbutton-click={(evt) => {
                action("subbutton-click")(evt.detail);
              }}
            >
              <SubButton value={"delta"} tooltips={deltaTooltips}>
                <Delta />%
              </SubButton>
              <SubButton value={"pie"} tooltips={pieTooltips}>
                <PieChart />%
              </SubButton>
            </ButtonGroup>
          </td>
        {/each}
      </tr>
    {/each}
  </table>
</Story>

<Story name="Button group, 2 sub-buttons">
  <ButtonGroup
    on:subbutton-click={(evt) => {
      action("subbutton-click")(evt.detail);
    }}
  >
    <SubButton value={1} tooltips={deltaTooltips}>
      <Delta />%
    </SubButton>
    <SubButton value={2} tooltips={pieTooltips}>
      <PieChart />%
    </SubButton>
  </ButtonGroup>
</Story>

<Story name="Button group, 4 sub-buttons">
  <ButtonGroup
    on:subbutton-click={(evt) => {
      action("subbutton-click")(evt.detail);
    }}
  >
    <SubButton value={1}>
      <Delta />%
    </SubButton>
    <SubButton value={2}>
      <PieChart />%
    </SubButton>
    <SubButton value={3}>
      <PieChart />%
    </SubButton>
    <SubButton value={4}>
      <PieChart />%
    </SubButton>
  </ButtonGroup>
</Story>

<style>
  td {
    padding: 10px;
  }
  td:first-child {
    padding-right: 40px;
  }
</style>
