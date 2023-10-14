export type ReportRun = {
  id: string;
  timestamp: number;
  exportSize: number;
  status: string; // success, failed, etc.
};

export const defaultData: ReportRun[] = [
  {
    id: "firstReport",
    timestamp: 1696555461000 - 800000,
    exportSize: 100000,
    status: "success",
  },
  {
    id: "thirdReport",
    timestamp: 1696555461000,
    exportSize: 200000,
    status: "success",
  },
  {
    id: "secondReport",
    timestamp: 1696555461000 - 200000,
    exportSize: 300000,
    status: "success",
  },
];
