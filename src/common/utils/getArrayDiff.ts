export function getArrayDiff<Src, Target>(
  srcArray: Array<Src>,
  srcGetter: (src: Src) => string,
  targetArray: Array<Target>,
  targetGetter: (target: Target) => string
) {
  const srcSet = new Set<string>();
  const extraSrc = new Array<Src>();
  srcArray.forEach((src) => srcSet.add(srcGetter(src)));

  const targetSet = new Set<string>();
  const extraTarget = new Array<Target>();
  targetArray.forEach((target) => {
    const val = targetGetter(target);
    targetSet.add(val);
    if (!srcSet.has(val)) {
      extraTarget.push(target);
    }
  });

  srcArray.forEach((src) => {
    const val = srcGetter(src);
    if (!targetSet.has(val)) {
      extraSrc.push(src);
    }
  });

  return {
    extraSrc,
    extraTarget,
  };
}
