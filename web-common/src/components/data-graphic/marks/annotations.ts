import type {
  GraphicScale,
  SimpleDataGraphicConfiguration,
} from "@rilldata/web-common/components/data-graphic/state/types";

export type Annotation = {
  time: Date;
  time_end?: Date;
  grain?: string;
  description: string;
};

export type AnnotationGroup = {
  items: Annotation[];
  top: number;
  left: number;
  bottom: number;
  right: number;
  hasRange: boolean;
};

// h-[24px] w-[12px]
export const AnnotationWidth = 12;
const AnnotationOverlapWidth = AnnotationWidth * (1 - 0.66); // Width where 66% overlap
export const AnnotationHeight = 12;

export function createAnnotationGroups(
  annotations: Annotation[],
  scaler: GraphicScale,
  config: SimpleDataGraphicConfiguration,
): AnnotationGroup[] {
  if (annotations.length === 0 || !scaler || !config) return [];

  const annotationTop = config.plotBottom - AnnotationHeight;
  const annotationBottom = config.plotBottom;

  const firstAnnotation = annotations[0];
  const firstAnnotationLeft = config.bodyLeft + scaler(firstAnnotation.time);
  const firstAnnotationRight =
    config.bodyLeft +
    (firstAnnotation.time_end
      ? scaler(firstAnnotation.time_end)
      : firstAnnotationLeft + AnnotationWidth);
  let currentGroup: AnnotationGroup = {
    items: [firstAnnotation],
    top: annotationTop,
    left: firstAnnotationLeft,
    bottom: annotationBottom,
    right: firstAnnotationRight,
    hasRange: !!firstAnnotation.time_end,
  };
  const groups: AnnotationGroup[] = [currentGroup];

  for (let i = 1; i < annotations.length; i++) {
    const annotation = annotations[i];
    const left = config.bodyLeft + scaler(annotation.time);
    const right =
      config.bodyLeft +
      (annotation.time_end
        ? scaler(annotation.time_end)
        : left + AnnotationWidth);

    const leftDiff = left - currentGroup.left;

    if (leftDiff < AnnotationOverlapWidth) {
      currentGroup.items.push(annotation);
    } else {
      currentGroup = {
        items: [annotation],
        top: annotationTop,
        left,
        right: Math.max(currentGroup.right, right),
        bottom: annotationBottom,
        hasRange: Boolean(currentGroup.hasRange || annotation.time_end),
      };
      groups.push(currentGroup);
    }
  }

  return groups;
}

export function buildLookupTable(annotationGroups: AnnotationGroup[]) {
  if (annotationGroups.length === 0) return [];
  const lastGroup = annotationGroups[annotationGroups.length - 1];

  const lookupTable = new Array<AnnotationGroup | undefined>(
    Math.ceil(lastGroup.right) + 1,
  ).fill(undefined);

  annotationGroups.forEach((group) => {
    const left = Math.floor(group.left);
    for (let x = 0; x <= AnnotationWidth; x++) {
      lookupTable[left + x] = group;
    }
  });

  return lookupTable;
}
