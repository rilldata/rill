export interface QueryError {
  response: {
    status: number;
    data: {
      message: string;
    };
  };
  message: string;
}
