<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import { Meta, Story, Template } from "@storybook/addon-svelte-csf";
  import Button from "../../button/Button.svelte";
  import YAMLEditor from "../YAMLEditor.svelte";
  import { setLineStatuses } from "../line-status";
  import type { LineStatus } from "../line-status/state";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifacts";

  let content = `name: this is the name
values:
  - label: another
    expression: count(*)
  - label: yet another
    expression: sum(revenue) / count(*)
  `;

  const fileArtifact = new FileArtifact("/file.yaml");

  let view: EditorView;

  let errors: LineStatus[] = [];
  function toggleError() {
    if (errors.length > 1) {
      errors = [];
    } else
      errors = [
        {
          line: 1,
          message:
            "This is an error that will always appear on line 1 even if you change the content",
          level: "error",
        },
      ];

    setLineStatuses(errors, view);
  }
</script>

<Meta title="Editor Components" />

<Template>
  <section class="space-y-3">
    <h1>Generic YAML editor</h1>
    <p class="w-96 ui-copy">
      This component can be used to edit any YAML content. It utilizes a set of
      CodeMirror plugins to manage line statuses and indent guides.
    </p>
    <div class="pb-3">
      <Button type="secondary" on:click={toggleError}
        >{#if errors?.length}Hide the line error{:else}Show the line error{/if}</Button
      >
    </div>
    <YAMLEditor
      {fileArtifact}
      key="key"
      {content}
      bind:view
      on:save={(event) => {
        // Often, you want to debounce the update to parent content.
        // Here, we have no such requirement.
        content = event.detail;
      }}
    />
  </section>
</Template>

<Story name="Generic YAML editor" />
