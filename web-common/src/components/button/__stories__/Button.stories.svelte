<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import Button from "@rilldata/web-common/components/button/Button.svelte";

  type ButtonTypes = "primary" | "secondary" | "text";
  type ButtonStatuses = "info" | "error";
  type ButtonProps = {
    type: ButtonTypes;
    status: ButtonStatuses;
    disabled: boolean;
    compact: boolean;
    label: string;
  };

  const buttonProps: ButtonProps[] = [];

  for (const buttonType of ["primary", "secondary", "text"]) {
    for (const status of ["info", "error"]) {
      for (const disabled of [true, false]) {
        for (const compact of [true, false]) {
          buttonProps.push({
            type: buttonType as ButtonTypes,
            status: status as ButtonStatuses,
            disabled,
            compact,
            label: `${buttonType} ${status} ${disabled} ${compact}`,
          });
        }
      }
    }
  }
</script>

<Meta title="Button stories" />

<Template>
  <table>
    <tr>
      <td />
      <td>type</td>
      <td>status</td>
      <td>disabled</td>
      <td>compact</td>
    </tr>
    {#each buttonProps as props}
      <tr>
        <td><Button {...props}>Button</Button></td>
        <td>{props.type}</td>
        <td>{props.status}</td>
        <td>{props.disabled}</td>
        <td>{props.compact}</td>
      </tr>
    {/each}
  </table>
</Template>

<Story name="all button variations" />

<style>
  td {
    padding: 5px;
  }
  td:first-child {
    padding-right: 40px;
  }
</style>
