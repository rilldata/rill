import { Annotation } from "@codemirror/state";
/** use this annotation to denote that the runtime has updated the content
 * of an editor buffer.
 */
export const outsideContentUpdateAnnotation = Annotation.define<string>();

export const debounceDocUpdateAnnotation = Annotation.define<number>();
