import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";

export function formatDataSizeQuota(
  storageLimitBytesPerDeployment: string,
): string {
  if (
    Number.isNaN(Number(storageLimitBytesPerDeployment)) ||
    storageLimitBytesPerDeployment === "-1"
  )
    return "";
  return `Max ${formatMemorySize(Number(storageLimitBytesPerDeployment))} / Project`;
}
