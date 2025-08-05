export function getUserFriendlyError(errorMessage: string): string {
  // Redshift
  if (errorMessage.includes("ERROR: UNLOAD destination is not supported")) {
    return `failed to unload: Redshift query execution failed ERROR: UNLOAD destination is not supported. (Hint: add output_location, role_arn, and database to your model)`;
  }
  if (errorMessage.includes("ERROR: AWS IAM role cannot be empty.")) {
    return `failed to unload: Redshift query execution failed ERROR: AWS IAM role cannot be empty. (Hint: add role_arn to your model)`;
  }

  return errorMessage;
}
