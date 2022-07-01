export const BATCH_SIZE = 1000;

export const PUBLISHER_DOMAINS = [
  ["Yahoo", "sports.yahoo.com"],
  ["Yahoo", "news.yahoo.com"],
  ["Microsoft", "msn.com"],
  ["Google", "news.google.com"],
  ["Google", "google.com"],
  ["Facebook", "facebook.com"],
  ["Facebook", "instagram.com"],
];
export const PUBLISHER_DOMAINS_BY_MONTH = [
  PUBLISHER_DOMAINS.filter((_, index) => index % 3 === 0),
  PUBLISHER_DOMAINS.filter((_, index) => index % 3 === 1),
  PUBLISHER_DOMAINS.filter((_, index) => index % 3 === 2),
];
export const START_DATE = "2022-01-01";
export const END_DATE = "2022-03-31";
export const BID_START = 1;
export const PUBLISHER_NULL_CHANCE = 0.33;

export const BID_END = 2;

export const LOCATIONS = [
  ["Bengaluru", "India"],
  ["Mumbai", "India"],
  ["Delhi", "India"],
  ["Kolkata", "India"],
  ["San Francisco", "USA"],
  ["Los Angeles", "USA"],
  ["New York City", "USA"],
  ["Boston", "USA"],
  ["Dublin", "Ireland"],
  ["London", "UK"],
];
export const CITY_NULL_CHANCE = 0.2;
export const USER_NULL_CHANCE = 0.5;

export const MAX_USERS = 100;
export const AD_BID_COUNT = 100000;
export const AD_IMPRESSION_COUNT = 50000;

export const DATA_FOLDER = "test/data";
