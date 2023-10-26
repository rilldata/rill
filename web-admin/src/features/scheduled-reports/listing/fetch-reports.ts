export type Report = {
  id: string;
  name: string;
  dashboard: string;
  author: string;
  dimensions: string[];
  metrics: string[];
  frequency: string; // daily, cronjob, etc.
  destination: string; // email, Slack, etc.
  lastRun: number;
  status: string; // success, failed, etc.
};

export const defaultData: Report[] = [
  {
    id: "firstReport",
    name: "First report",
    dashboard: "Marketing dashboard",
    author: "Eric",
    dimensions: ["state", "channel"],
    metrics: ["impressions", "signups"],
    frequency: "daily",
    destination: "email",
    lastRun: 1696555461000 - 800000,
    status: "success",
  },
  {
    id: "thirdReport",
    name: "Third report",
    dashboard: "Operations dashboard",
    author: "Janet",
    dimensions: ["state", "channel"],
    metrics: ["impressions", "signups"],
    frequency: "daily",
    destination: "email",
    lastRun: 1696555461000,
    status: "success",
  },
  {
    id: "secondReport",
    name: "Second report",
    dashboard: "Engineering dashboard",
    author: "John Doe",
    dimensions: ["state", "channel"],
    metrics: ["impressions", "signups"],
    frequency: "daily",
    destination: "email",
    lastRun: 1696555461000 - 200000,
    status: "success",
  },
];
