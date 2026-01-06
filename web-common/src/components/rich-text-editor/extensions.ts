import StarterKit from "@tiptap/starter-kit";
import { Placeholder } from "@tiptap/extensions";
import { MeasureMention } from "./measure-mention";
import { RepeatBlock } from "./repeat-block";
import MeasureMentionPicker from "./MeasureMentionPicker.svelte";

export function getEditorExtensions({
  placeholder,
  metricsViewName,
  availableMeasures,
}: {
  placeholder: string;
  metricsViewName?: string;
  availableMeasures?: string[];
}) {
  const extensions = [
    StarterKit.configure({
      heading: {
        levels: [1, 2, 3, 4, 5, 6],
      },
    }),
    Placeholder.configure({
      placeholder,
    }),
  ];

  // Add measure mention extension if we have a metrics view
  // Always add it if we have a metricsViewName, even if measures aren't loaded yet
  if (metricsViewName) {
    // Store measures in a variable that the closures can access
    const measuresList = availableMeasures || [];

    extensions.push(
      MeasureMention.configure({
        metricsViewName,
        availableMeasures: measuresList,
        suggestion: {
          char: "@",
          allowSpaces: false,
          items: ({ query }) => {
            // Use the closure variable which captures the current value
            const measures = measuresList;
            
            if (measures.length === 0) {
              return [];
            }

            const filtered = measures
              .filter((measure) =>
                measure.toLowerCase().includes(query.toLowerCase()),
              )
              .slice(0, 20)
              .map((measure) => ({
                id: measure,
                label: measure,
              }));
            return filtered;
          },
          render: () => {
            let pickerComponent: any = null;
            let selected = false;

            return {
              onStart: (props: any) => {
                if (!(props.decorationNode instanceof HTMLElement)) return;
                selected = false;

                // Use the closure variable which captures the current value
                const measures = measuresList;

                if (measures.length === 0) {
                  // Don't show picker if no measures available
                  return;
                }

                pickerComponent = new MeasureMentionPicker({
                  target: document.body,
                  props: {
                    availableMeasures: measures,
                    searchText: props.query || "",
                    refNode: props.decorationNode,
                    onSelect: (measure: string) => {
                      selected = true;
                      // Find the item index and select it
                      const itemIndex = props.items.findIndex(
                        (item: any) => item.id === measure,
                      );
                      if (itemIndex >= 0) {
                        props.selectItem(itemIndex);
                      }
                    },
                    focusEditor: () => props.editor.commands.focus(),
                  },
                });
              },
              onUpdate: (props: any) => {
                if (!(props.decorationNode instanceof HTMLElement)) return;
                
                // Use the closure variable which captures the current value
                const measures = measuresList;

                if (pickerComponent && measures.length > 0) {
                  pickerComponent.$set({
                    availableMeasures: measures,
                    searchText: props.query || "",
                    refNode: props.decorationNode,
                  });
                }
              },
              onExit: () => {
                if (pickerComponent) {
                  pickerComponent.$destroy();
                  pickerComponent = null;
                }
              },
            };
          },
        },
      }),
    );
  }

  // Add repeat block extension
  extensions.push(RepeatBlock);

  return extensions;
}

