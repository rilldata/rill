import Mention, { type MentionOptions } from "@tiptap/extension-mention";
import { Extension } from "@tiptap/core";
import InlineContextPicker from "@rilldata/web-common/features/chat/core/context/picker/InlineContextPicker.svelte";
import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import InlineContextComponent from "@rilldata/web-common/features/chat/core/context/InlineContext.svelte";
import type { EditorView } from "@tiptap/pm/view";
import { PluginKey, type EditorState } from "@tiptap/pm/state";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { Placeholder, UndoRedo } from "@tiptap/extensions";
import { mount, unmount, getAllContexts } from "svelte";
import {
  convertContextToInlinePrompt,
  INLINE_CHAT_CONTEXT_TAG,
  type InlineContext,
  normalizeInlineContext,
  parseInlineAttr,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { PickerOptionsGetter } from "@rilldata/web-common/features/chat/core/context/picker/filters.ts";

export function getEditorPlugins({
  placeholder,
  onSubmit,
  skillOptions,
}: {
  placeholder: string;
  onSubmit: () => void;
  // Adds a "/" picker listing the skills this getter returns.
  skillOptions?: PickerOptionsGetter;
}) {
  const sharedEditorStore = new SharedEditorStore();

  const plugins = [
    Document,
    Paragraph,
    Text,
    Placeholder.configure({
      placeholder,
    }),
    EditorSubmitExtension.configure({ onSubmit, sharedEditorStore }),
    configureInlineContextTipTapExtension(sharedEditorStore, skillOptions),
    UndoRedo,
  ];

  return plugins;
}

/**
 * Hooks into the editor's shortcut system.
 * Maps Shift-Enter to the editor's enter command.
 * Maps Enter to the submit action calling the onSubmit callback.
 * Also suppresses up and down arrow keys when context picker is open.
 */
const EditorSubmitExtension = Extension.create(() => {
  let isShiftEnter = false;

  return {
    name: "editorSubmit",

    addOptions() {
      return {
        onSubmit: () => {},
        sharedEditorStore: <SharedEditorStore>{},
      };
    },

    addKeyboardShortcuts() {
      return {
        Enter: () => {
          if (!isShiftEnter) {
            // Suppress enter to submit when context picker is open
            if (!this.options.sharedEditorStore.contextOpen) {
              this.options.onSubmit?.();
            }
            return true;
          }
          isShiftEnter = false;
          return false;
        },
        "Shift-Enter": () => {
          isShiftEnter = true;
          return this.editor.commands.enter();
        },
        // Suppress arrow keys when context picker is open
        ArrowDown: () => {
          return this.options.sharedEditorStore.contextOpen;
        },
        ArrowUp: () => {
          return this.options.sharedEditorStore.contextOpen;
        },
        ArrowLeft: () => {
          return this.options.sharedEditorStore.contextOpen;
        },
        ArrowRight: () => {
          return this.options.sharedEditorStore.contextOpen;
        },
      };
    },
  };
});

type InlineContextOptions = MentionOptions<never, InlineContext> & {
  sharedEditorStore: SharedEditorStore;
  allParentContexts: Map<any, any>;
};

// Add the startMention and startSkill commands to the Commands type.
declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    mention: {
      startMention: () => ReturnType;
      startSkill: () => ReturnType;
    };
  }
}

/**
 * Extends the existing Mention extension to support inline chat context.
 * Creates InlineChatContext svelte component to display an interactive inline chat context block.
 */
const InlineContextExtension = Mention.extend<InlineContextOptions>({
  // Add a param for ConversationManager on top of Mention's options.
  addOptions() {
    return {
      ...((this.parent?.() ?? {}) as MentionOptions<never, InlineContext>),
      // These have to be configured for the extension to work
      manager: {} as ConversationManager,
      sharedEditorStore: {} as SharedEditorStore,
      allParentContexts: new Map(),
    };
  },

  // Mapping for attributes. We need to map values in InlineChatContext to html attribute and vice-versa.
  addAttributes() {
    return {
      type: createAttributeEntry(null, "type"),
      metricsView: createAttributeEntry(null, "metricsView"),
      canvas: createAttributeEntry(null, "canvas"),
      canvasComponent: createAttributeEntry(null, "canvasComponent"),
      measure: createAttributeEntry(null, "measure"),
      dimension: createAttributeEntry(null, "dimension"),
      timeRange: createAttributeEntry(null, "timeRange"),
      filePath: createAttributeEntry(null, "filePath"),
      model: createAttributeEntry(null, "model"),
      column: createAttributeEntry(null, "column"),
      columnType: createAttributeEntry(null, "columnType"),
      skill: createAttributeEntry(null, "skill"),
    };
  },

  addCommands() {
    return {
      startMention:
        () =>
        ({ tr, view, commands }) => {
          commands.focus();
          // Only focus the editor if context is already open.
          if (this.options.sharedEditorStore.contextOpen) return false;

          tr.insertText("@");
          view.dispatchEvent(new KeyboardEvent("keyup", { key: "@" }));
          return true;
        },
      startSkill:
        () =>
        ({ tr, view, commands }) => {
          commands.focus();
          // Only focus the editor if context is already open.
          if (this.options.sharedEditorStore.contextOpen) return false;

          // The picker only opens at the start of a line or after a space, so add one when typing right after a word.
          const before = tr.doc.textBetween(
            Math.max(tr.selection.from - 1, 0),
            tr.selection.from,
          );
          tr.insertText(before && before !== " " ? " /" : "/");
          view.dispatchEvent(new KeyboardEvent("keyup", { key: "/" }));
          return true;
        },
    };
  },

  parseHTML() {
    return [
      {
        tag: INLINE_CHAT_CONTEXT_TAG,
      },
    ];
  },

  renderHTML({ HTMLAttributes }) {
    return [INLINE_CHAT_CONTEXT_TAG, HTMLAttributes, ""];
  },

  renderText({ node }) {
    return convertContextToInlinePrompt(node.attrs as InlineContext);
  },

  addNodeView() {
    return ({ node, getPos, view, editor }) => {
      // Create a wrapper div to render the component.
      // We need this since svelte only takes a target wrapper.
      const target = document.createElement("div");
      // We need this here to make sure the component is rendered inline.
      target.className = "inline-block";

      const { sharedEditorStore, allParentContexts } = this.options;

      // Create the inline chat context component. Pass the wrapper as the target.
      const comp = mount(InlineContextComponent, {
        target,
        props: {
          selectedChatContext: normalizeInlineContext(
            node.attrs as InlineContext,
          ),
          props: editor.options.editable
            ? {
                mode: "editable",
                onSelect: (selectedChatContext: InlineContext) => {
                  const pos = getPos();
                  if (!pos) return;

                  // Dispatch a transaction to update the node attributes with the new context.
                  view.dispatch(
                    getTransactionForContext(selectedChatContext, view, pos),
                  );
                  editor.commands.focus();
                },
                onDropdownToggle: (isOpen: boolean) =>
                  sharedEditorStore.dropdownToggled(comp, isOpen),
                focusEditor: () => editor.commands.focus(),
              }
            : { mode: "readonly" },
        },
        context: allParentContexts,
      }) as InlineContextExports;
      sharedEditorStore.componentAdded(comp);

      return {
        dom: target,
        destroy() {
          sharedEditorStore.componentsRemoved(comp);
          unmount(comp);
        },
      };
    };
  },
});

/**
 * Configures the InlineContextExtension to show a dropdown when the user types "@".
 * With `skillOptions`, also shows a dropdown listing skills when the user types "/".
 * Renders the InlineContextPicker svelte component.
 */
export function configureInlineContextTipTapExtension(
  sharedEditorStore: SharedEditorStore,
  skillOptions?: PickerOptionsGetter,
) {
  const allParentContexts = getAllContexts();

  const suggestions = [
    getPickerSuggestion(sharedEditorStore, allParentContexts, {
      char: "@",
      allowSpaces: true,
    }),
  ];
  if (skillOptions) {
    suggestions.push(
      getPickerSuggestion(sharedEditorStore, allParentContexts, {
        char: "/",
        getOptions: skillOptions,
        // A prompt references one skill at a time.
        multiple: false,
      }),
    );
  }

  // Only one picker can be open at a time. "@orders /mon" matches both "@" (spaces are allowed) and "/",
  // so a trigger is not allowed while the suggestion of another trigger is active.
  const pluginKeys = suggestions.map(
    () => new PluginKey<{ active: boolean }>(),
  );
  suggestions.forEach((suggestion, i) => {
    suggestion.pluginKey = pluginKeys[i];
    suggestion.allow = ({ editor, state, range }) =>
      isMentionAllowedAt(state, range.from) &&
      pluginKeys.every(
        (key, j) =>
          i === j ||
          // The state of a later plugin is not computed yet in the new state, so fall back to the current one.
          !(key.getState(state) ?? key.getState(editor.state))?.active,
      );
  });

  return InlineContextExtension.configure({
    sharedEditorStore,
    allParentContexts,
    suggestions,
  });
}

// Mention's default `allow`, which is replaced when `allow` is set.
function isMentionAllowedAt(state: EditorState, pos: number) {
  const type = state.schema.nodes[InlineContextExtension.name];
  return !!state.doc.resolve(pos).parent.type.contentMatch.matchType(type);
}

function getPickerSuggestion(
  sharedEditorStore: SharedEditorStore,
  allParentContexts: InlineContextOptions["allParentContexts"],
  {
    char,
    allowSpaces = false,
    getOptions,
    multiple = true,
  }: {
    char: string;
    allowSpaces?: boolean;
    getOptions?: PickerOptionsGetter;
    multiple?: boolean;
  },
): InlineContextOptions["suggestion"] {
  let comp: Record<string, unknown> | null = null;
  const pickerProps: Record<string, unknown> = $state({ getOptions, multiple });
  let selected = false;

  return {
    char,
    allowSpaces,
    items: () => [], // TODO: would it make sense to manage the options here?
    render: () => ({
      onStart: (props) => {
        if (!(props.decorationNode instanceof HTMLElement)) return; // type safety, non-html will be in non-dom environment
        selected = false;

        pickerProps.refNode = props.decorationNode;
        pickerProps.onSelect = (item: InlineContext) => {
          selected = true;
          props.command(item);
        };
        pickerProps.focusEditor = () => props.editor.commands.focus();
        comp = mount(InlineContextPicker, {
          target: document.body,
          props: pickerProps,
          context: allParentContexts,
        });
        sharedEditorStore.contextOpen = true;
      },

      onUpdate(props) {
        if (!(props.decorationNode instanceof HTMLElement)) return; // type safety, non-html will be in non-dom environment
        pickerProps.searchText = props.query;
        pickerProps.refNode = props.decorationNode;
      },

      onExit: ({ editor, range }) => {
        if (!comp) return;
        unmount(comp);
        comp = null;
        sharedEditorStore.contextOpen = false;

        if (!selected) return;
        // Remove the query text and replace with space.
        // This is not automatically removed by tiptap
        editor.view.dispatch(
          editor.view.state.tr.replaceRangeWith(
            range.from + 1,
            range.to + 1,
            editor.state.schema.text(" "),
          ),
        );
      },
    }),
  };
}

type InlineContextExports = { closeDropdown: () => void };

/**
 * Used to share data across plugins of editor.
 * It is used to keep track of the state of the context picker dropdowns.
 * It also keeps track of the components that are currently rendered and makes sure only one dropdown is open at a time.
 */
class SharedEditorStore {
  public contextOpen: boolean = false;
  private components: InlineContextExports[] = [];

  public componentAdded(comp: InlineContextExports) {
    this.components.push(comp);
  }

  public componentsRemoved(comp: InlineContextExports) {
    this.components = this.components.filter((c) => c !== comp);
    this.contextOpen = false;
  }

  public dropdownToggled(comp: InlineContextExports, isOpen: boolean) {
    this.contextOpen = isOpen;
    if (!isOpen) return;

    // If the dropdown for the current component was opened, close dropdowns for all other components.
    this.components.forEach((c) => {
      if (c === comp) return;
      c.closeDropdown();
    });
  }
}

function getTransactionForContext(
  inlineChatContext: InlineContext,
  view: EditorView,
  pos: number,
) {
  return view.state.tr
    .setNodeAttribute(pos, "type", inlineChatContext.type)
    .setNodeAttribute(pos, "metricsView", inlineChatContext.metricsView)
    .setNodeAttribute(pos, "canvas", inlineChatContext.canvas)
    .setNodeAttribute(pos, "canvasComponent", inlineChatContext.canvasComponent)
    .setNodeAttribute(pos, "measure", inlineChatContext.measure)
    .setNodeAttribute(pos, "dimension", inlineChatContext.dimension)
    .setNodeAttribute(pos, "timeRange", inlineChatContext.timeRange)
    .setNodeAttribute(pos, "model", inlineChatContext.model)
    .setNodeAttribute(pos, "column", inlineChatContext.column)
    .setNodeAttribute(pos, "columnType", inlineChatContext.columnType)
    .setNodeAttribute(pos, "skill", inlineChatContext.skill);
}

function createAttributeEntry(defaultValue: string | null, key: string) {
  return {
    default: defaultValue,
    parseHTML: (element: HTMLElement) =>
      element.getAttribute(key) ?? // Parsing from html attribute.
      parseInlineAttr(element.innerHTML, key) ?? // Parsing from inline prompt.
      defaultValue,
  };
}
