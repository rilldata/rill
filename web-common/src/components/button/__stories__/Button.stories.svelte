<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import Button from "@rilldata/web-common/components/button/Button.svelte";

  type ButtonProps = {
    type: "primary" | "secondary" | "text";
    status: "info" | "error";
    disabled: boolean;
    compact: boolean;
    label: string;
  };

  const buttonProps: ButtonProps[] = [];

  for (const type of ["primary", "secondary", "text"]) {
    for (const status of ["info", "error"]) {
      for (const disabled of [true, false]) {
        for (const compact of [true, false]) {
          buttonProps.push({
            type,
            status,
            disabled,
            compact,
            label: `${type} ${status} ${disabled} ${compact}`,
          });
        }
      }
    }
  }
</script>

<Meta title="Button stories" />

<Template let:args>
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
