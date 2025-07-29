import type { ScaleTime } from "d3-scale";

export type Annotation = {
  time: Date;
  time_end?: Date;
  grain?: string;
  description: string;
};

export type AnnotationGroup = {
  items: Annotation[];
  left: number;
  // right: number;
};

// h-[24px] w-[12px]
export const AnnotationWidth = 12;
const AnnotationOverlapWidth = AnnotationWidth * (1 - 0.66); // Width where 66% overlap
export const AnnotationHeight = 24;

export function createAnnotationGroups(
  annotations: Annotation[],
  scaler: ScaleTime<Date, number>,
): AnnotationGroup[] {
  if (annotations.length === 0) return [];

  let currentGroup: AnnotationGroup = {
    items: [annotations[0]],
    left: scaler(new Date(annotations[0].time)),
  };
  const groups: AnnotationGroup[] = [currentGroup];

  for (let i = 1; i < annotations.length; i++) {
    const annotation = annotations[i];
    const left = scaler(annotation.time);

    const leftDiff = left - currentGroup.left;

    if (leftDiff < AnnotationOverlapWidth) {
      currentGroup.items.push(annotation);
    } else {
      currentGroup = {
        items: [annotation],
        left,
      };
      groups.push(currentGroup);
    }
  }

  return groups;
}
