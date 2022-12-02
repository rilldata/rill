export function getColumn(profileColumns, columnName) {
  return profileColumns?.data?.profileColumns?.find(
    (column) => column.name === columnName
  );
}
