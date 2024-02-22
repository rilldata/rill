<script lang="ts">
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";

  import Button, {
    ButtonShape,
    ButtonSize,
    ButtonKind,
  } from "@rilldata/web-common/components/button/Button.svelte";

  // export const meta = {
  //   title: "Button stories",
  // };

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

  // outermost table layer

  type FigmaRow = {
    shape: ButtonShape;
    size: ButtonSize;
    danger: boolean;
    state: string;
  };

  const shapes: ButtonShape[] = ["normal", "square", "circle"];
  const sizes: ButtonSize[] = ["xl", "large", "medium", "small"];
  const states = ["normal", "hovered/pressed", "active", "status", "disabled"];

  const figmaRows: FigmaRow[] = [];

  for (const shape of shapes) {
    const dangers = shape === "normal" ? [false, true] : [false];
    for (const danger of dangers) {
      for (const size of sizes) {
        for (const state of states) {
          figmaRows.push({
            shape,
            size,
            danger,
            state,
          });
        }
      }
    }
  }

  // used as columns
  const buttonTypes: ButtonKind[] = [
    "primary",
    "secondary",
    "subtle",
    "ghost",
    "dashed",
    "link",
    "text",
  ];
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

<Story name="Button table (like Figma)">
  <p>
    This story lays out the available button types in a table that is intended
    to match the
    <a
      href="https://www.figma.com/file/nqqazRo1ckU9ooC9ym9weI/Rill-Design-System?type=design&node-id=1155%3A311096&mode=design&t=Ux60sppUPewJKbIW-1"
    >
      mocks in our design system in Figma</a
    > (as of Feb 2024).
  </p>
  <br />
  <!-- {#each figmaRows as row}
    <h1>{shape}</h1>
  {/each} -->

  <table>
    <thead>
      <tr class="red">
        <th></th>
        {#each buttonTypes as buttonType}
          <th><h2>{buttonType}</h2></th>
        {/each}
      </tr>
    </thead>

    {#each figmaRows as row}
      <tr>
        <td>
          {#if row.size === "xl" && row.state === "normal"}
            <h1>{row.danger ? "dangerous" : row.shape}</h1>
            <!-- {:else} -->
          {/if}
          {#if row.state === "normal"}
            <h2>{row.size}</h2>
            <!-- {:else} -->
          {/if}
          {row.state}
          <!-- <br />{JSON.stringify(row)} -->
        </td>
        {#each buttonTypes as buttonType}
          <td>
            <Button
              shape={row.shape}
              size={row.size}
              status={row.danger ? "error" : "info"}
              type={buttonType}
              selected={row.state === "hovered/pressed"}
              active={row.state === "active"}
              disabled={row.state === "disabled"}
              loading={row.state === "status"}
              >{row.shape === "normal" ? "Button Title" : "A"}</Button
            ></td
          >
        {/each}
      </tr>
    {/each}
  </table>
</Story>

<style>
  td {
    padding: 5px;
    vertical-align: bottom;
    text-align: center;
  }
  td:first-child {
    padding-right: 40px;
  }

  th {
    background: white;
    position: sticky;
    top: 0; /* Don't forget this, required for the stickiness */
    box-shadow: 0 2px 2px -1px rgba(0, 0, 0, 0.4);
  }

  h1 {
    font-size: 24px;
    padding-top: 20px;
  }
  h2 {
    font-size: 18px;
    color: #999;
    padding-top: 20px;
  }
</style>
