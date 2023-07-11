import { Annotation } from "@codemirror/state";

/** CodeMirror annotation that provides a boolean the doc update
 * itself should skip debouncing. This is used in situations where an immediate
 * reconciliation is necessary, for instance an action to update a YAML file
 * through a non-editing user action (such as "template in this file").
 */
export const skipDebounceAnnotation = Annotation.define<boolean>();
