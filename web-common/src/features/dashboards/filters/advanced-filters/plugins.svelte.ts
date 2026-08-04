import { Mention, type MentionOptions } from "@tiptap/extension-mention";
import {
  ADVANCED_FILTER_TAG,
  type AdvancedFilter,
  type AdvancedFilterMentionOption,
  convertAdvancedFilterToHTML,
} from "@rilldata/web-common/features/dashboards/filters/advanced-filters/advanced-filter.ts";
import { getAllContexts, mount, unmount } from "svelte";
import AdvancedFilterPill from "@rilldata/web-common/features/dashboards/filters/advanced-filters/AdvancedFilterPill.svelte";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import AddAdvancedFilter from "@rilldata/web-common/features/dashboards/filters/advanced-filters/AddAdvancedFilter.svelte";
import { Placeholder, UndoRedo } from "@tiptap/extensions";
import type { EditorView } from "@tiptap/pm/view";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { parseInlineAttr } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export function getAdvancedFilterEditorPlugins({
  expressionFilterManager,
  dimensions,
  measures,
}: {
  expressionFilterManager: ExpressionFilterManager;
  dimensions: MetricsViewSpecDimension[];
  measures: MetricsViewSpecMeasure[];
}) {
  const plugins = [
    Document,
    Paragraph,
    Text,
    Placeholder.configure({
      placeholder: "Enter advanced filter. Press @ for dimensions/measures",
    }),
    configureAdvancedFilterTipTapExtension(
      expressionFilterManager,
      dimensions,
      measures,
    ),
    UndoRedo,
  ];

  return plugins;
}

type FilterPillOptions = MentionOptions<
  AdvancedFilterMentionOption,
  AdvancedFilter
> & {
  expressionFilterManager: ExpressionFilterManager | undefined;
  allParentContexts: Map<any, any>;
};

const FilterPillPlugin = Mention.extend<FilterPillOptions>({
  addOptions() {
    return {
      ...((this.parent?.() ?? {}) as MentionOptions<never, AdvancedFilter>),
      expressionFilterManager: undefined,
      // These have to be configured for the extension to work
      allParentContexts: new Map(),
    };
  },

  // Mapping for attributes. We need to map values in InlineChatContext to html attribute and vice-versa.
  addAttributes() {
    return {
      name: createAttributeEntry("name"),
      sql: createAttributeEntry("sql"),
    };
  },

  parseHTML() {
    return [{ tag: ADVANCED_FILTER_TAG }];
  },

  renderHTML({ HTMLAttributes }) {
    return [ADVANCED_FILTER_TAG, HTMLAttributes, ""];
  },

  renderText({ node }) {
    return convertAdvancedFilterToHTML(node.attrs as AdvancedFilter);
  },

  addNodeView() {
    return ({ node, getPos, view, editor }) => {
      // Create a wrapper div to render the component.
      // We need this since svelte only takes a target wrapper.
      const target = document.createElement("div");
      // We need this here to make sure the component is rendered inline.
      target.className = "inline-block";

      const { allParentContexts, expressionFilterManager } = this.options;

      // Create the inline chat context component. Pass the wrapper as the target.
      const comp = mount(AdvancedFilterPill, {
        target,
        props: {
          advancedFilter: node.attrs as AdvancedFilter,
          expressionFilterManager,
          onChange: (advancedFilter: AdvancedFilter) => {
            const pos = getPos();
            if (!pos) return;

            // Dispatch a transaction to update the node attributes with the new context.
            view.dispatch(getTransactionForFilter(advancedFilter, view, pos));
            editor.commands.focus();
          },
          focusEditor: () => editor.commands.focus(),
        },
        context: allParentContexts,
      });

      return {
        dom: target,
        destroy() {
          unmount(comp);
        },
      };
    };
  },
});

export function configureAdvancedFilterTipTapExtension(
  expressionFilterManager: ExpressionFilterManager,
  dimensions: MetricsViewSpecDimension[],
  measures: MetricsViewSpecMeasure[],
) {
  let comp: Record<string, unknown> | null = null;
  const pickerProps: Record<string, unknown> = $state({});

  const allParentContexts = getAllContexts();

  return FilterPillPlugin.configure({
    expressionFilterManager,
    allParentContexts,
    suggestion: {
      char: "@",
      allowSpaces: true,
      items: () => [
        ...dimensions.map(
          (d) =>
            <AdvancedFilterMentionOption>{ type: "dimension", dimension: d },
        ),
        ...measures.map(
          (m) => <AdvancedFilterMentionOption>{ type: "measure", measure: m },
        ),
      ],
      render: () => ({
        onStart: (props) => {
          if (!(props.decorationNode instanceof HTMLElement)) return; // type safety, non-html will be in non-dom environment

          pickerProps.items = props.items;
          pickerProps.refNode = props.decorationNode;
          pickerProps.onSelect = (item: AdvancedFilter) => {
            props.command(item);
          };
          comp = mount(AddAdvancedFilter, {
            target: document.body,
            props: pickerProps,
            context: allParentContexts,
          });
        },

        onUpdate(props) {
          if (!(props.decorationNode instanceof HTMLElement)) return; // type safety, non-html will be in non-dom environment
          pickerProps.items = props.items;
          pickerProps.refNode = props.decorationNode;
        },

        onExit: () => {
          if (!comp) return;
          unmount(comp);
          comp = null;
        },
      }),
    },
  });
}

function getTransactionForFilter(
  advancedFilter: AdvancedFilter,
  view: EditorView,
  pos: number,
) {
  return view.state.tr
    .setNodeAttribute(pos, "sql", advancedFilter.sql)
    .setNodeAttribute(pos, "name", advancedFilter.name);
}

function createAttributeEntry(key: string) {
  return {
    default: null,
    parseHTML: (element: HTMLElement) =>
      element.getAttribute(key) ?? // Parsing from html attribute.
      parseInlineAttr(element.innerHTML, key) ?? // Parsing from inline prompt.
      null,
  };
}
