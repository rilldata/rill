export interface DomainCoordinates<T extends number | Date = number | Date> {
  x?: T;
  y?: number;
  // For annotations we need the actual x/y. So save it directly for easy access.
  xActual?: number;
  yActual?: number;
}
