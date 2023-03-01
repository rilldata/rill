const ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX = {
  0: "",
  3: "k",
  6: "M",
  9: "B",
  12: "T",
  15: "Q",
};

export const shortScaleSuffixIfAvailable = (x: number): string => {
  let suffix = ORDER_OF_MAG_TO_SHORT_SCALE_SUFFIX[x];
  if (suffix !== undefined) return suffix;
  return "E" + x;
};

const ORDER_OF_MAG_TEXT_TO_SHORT_SCALE_SUFFIX = {
  E0: "",
  E3: "k",
  E6: "M",
  E9: "B",
  E12: "T",
  E15: "Q",
};
export const shortScaleSuffixIfAvailableForStr = (suffixIn: string): string => {
  let suffix = ORDER_OF_MAG_TEXT_TO_SHORT_SCALE_SUFFIX[suffixIn.toUpperCase()];
  if (suffix !== undefined) return suffix;
  return suffixIn;
};
