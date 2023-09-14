// The Orval-generated type for query errors is `RpcStatus`, but the following is the observed type that is actually returned.
export interface QueryError {
  response: {
    status: number;
    data: {
      message: string;
    };
  };
  message: string;
}
