<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import SummaryStatistics from "@rilldata/web-common/components/icons/SummaryStatistics.svelte";

  type ButtonProps = {
    disabled: boolean;
    rounded: boolean;
    compact: boolean;
    bgDark: boolean;
    active: boolean;
  };

  const buttonProps: ButtonProps[] = [];

  for (const disabled of [true, false]) {
    for (const rounded of [true, false]) {
      for (const compact of [true, false]) {
        for (const bgDark of [true, false]) {
          for (const active of [true, false]) {
            buttonProps.push({
              disabled,
              rounded,
              compact,
              bgDark,
              active,
            });
          }
        }
      }
    }
  }
</script>

<Meta title="Icon Button stories" />

<Template let:args>
  <table>
    <tr>
      <td />
      <td>disabled</td>
      <td>rounded</td>
      <td>compact</td>
      <td>bgDark</td>
      <td>active</td>
    </tr>
    {#each buttonProps as props}
      <tr>
        <td
          ><IconButton {...props}
            ><SummaryStatistics />
            <svelte:fragment slot="tooltip-content">
              Show basic summary statistics
            </svelte:fragment></IconButton
          ></td
        >
        <td>{props.disabled}</td>
        <td>{props.rounded}</td>
        <td>{props.compact}</td>
        <td>{props.bgDark}</td>
        <td>{props.active}</td>
      </tr>
    {/each}
  </table>
</Template>

<Story name="all icon button variations" />

<style>
  td {
    padding: 5px;
  }
  td:first-child {
    padding-right: 40px;
  }
</style>
