export enum ConversationContextType {
  MetricsView,
  TimeRange,
  Measures,
}

export type ConversationContextEntry = {
  type: ConversationContextType;
  value: string;
};
