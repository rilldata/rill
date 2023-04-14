/**
 * Provides the state field & effect for line status. We use this to
 * place line statuses in the editor and gutter, such as inline errors.
 */
import { StateEffect, StateField } from "@codemirror/state";

/** defines a state effect that updates the lineState field. */
export const updateLineStatus = StateEffect.define<{
  lineState: Array<{ line: number; message: string; level: string }>;
}>({
  map: (value, mapping) => {
    return {
      lineState: value.lineState
        .filter((line) => line.line !== null)
        .map((line) => ({
          line: mapping.mapPos(line.line),
          message: line.message,
          level: line.level,
        })),
    };
  },
});

/** defines the line status state field, used to show different kinds of
 * ... line statuses, such as errors, warnings, info, etc.
 */
export const lineStatusStateField = StateField.define({
  create: () => [],
  update: (lines, tr) => {
    // Handle transactions with the updateLineState effect
    for (const effect of tr.effects) {
      if (effect.is(updateLineStatus)) {
        // Clear the existing errors and set the new errors
        return effect.value.lineState.slice();
      }
    }

    return lines;
  },
  compare: (a, b) => a === b,
});

export const levels = {
  error: {
    bgColor: "rgba(255,0,0,.1)",
  },
};
