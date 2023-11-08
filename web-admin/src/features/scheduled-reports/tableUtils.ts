export function formatDateToCustomString(date: Date) {
  const formattedDate = date.toLocaleString("en-US", {
    month: "short",
    day: "2-digit",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  });

  const dateParts = formattedDate.split(", ");
  dateParts[0] = dateParts[0] + " " + dateParts[1];
  dateParts.splice(1, 1);

  return dateParts.join(", ").replace(" PM", "pm").replace(" AM", "am");
}

export function capitalizeFirstLetter(string: string) {
  return string.charAt(0).toUpperCase() + string.slice(1);
}
