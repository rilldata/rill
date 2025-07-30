import type {
  GraphicScale,
  SimpleDataGraphicConfiguration,
} from "@rilldata/web-common/components/data-graphic/state/types";

export type Annotation = {
  startTime: Date;
  truncatedStartTime: Date;
  endTime?: Date;
  truncatedEndTime?: Date;
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
  rangeLeft: number;
  rangeRight: number;
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

  let currentGroup: AnnotationGroup = getSingletonAnnotationGroup(
    annotations[0],
    scaler,
    config,
  );
  const groups: AnnotationGroup[] = [currentGroup];

  for (let i = 1; i < annotations.length; i++) {
    const annotation = annotations[i];
    const group = getSingletonAnnotationGroup(annotation, scaler, config);

    const leftDiff = group.left - currentGroup.left;

    if (leftDiff < AnnotationOverlapWidth) {
      currentGroup.right = Math.max(currentGroup.right, group.right);
      currentGroup.rangeRight = Math.max(
        currentGroup.rangeRight,
        group.rangeRight,
      );
      currentGroup.items.push(annotation);
    } else {
      currentGroup = group;
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

function getSingletonAnnotationGroup(
  annotation: Annotation,
  scaler: GraphicScale,
  config: SimpleDataGraphicConfiguration,
): AnnotationGroup {
  const left = config.bodyLeft + scaler(annotation.startTime);
  const rangeLeft = config.bodyLeft + scaler(annotation.truncatedStartTime);
  const right =
    config.bodyLeft +
    (annotation.endTime ? scaler(annotation.endTime) : left + AnnotationWidth);
  const rangeRight =
    config.bodyLeft +
    (annotation.truncatedEndTime ? scaler(annotation.truncatedEndTime) : right);
  return <AnnotationGroup>{
    items: [annotation],
    top: config.plotBottom - AnnotationHeight + 3,
    left,
    rangeLeft,
    bottom: config.plotBottom,
    right,
    rangeRight,
    hasRange: !!annotation.endTime,
  };
}
