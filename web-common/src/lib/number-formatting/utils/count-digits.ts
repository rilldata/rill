export const countDigits = (numStr: string) => {
  if (numStr === "0") return 0;
  return numStr.replace(/[^0-9]/g, "").length;
};

export const countNonZeroDigits = (numStr: string) =>
  numStr.replace(/[^1-9]/g, "").length;
