WITH  

SanFranciscoNormalized AS 
(
SELECT
  'San Francisco' AS city,
  'CA' AS state,
  Opened AS start_event_date,
  STRPTIME(Closed, '%m/%d/%Y %I:%M:%S %p') AS end_event_date,
  CASE WHEN Longitude = '0.0' THEN NULL ELSE Longitude END AS longitude,
  CASE WHEN Latitude = '0.0' THEN NULL ELSE Latitude END  AS latitude,
  CaseID AS ticket_id,
  SPLIT(Address, ', ')[1] AS street_address,
  Street AS street,
  SPLIT(Address, ', ')[-1]  AS zipcode,
  CONCAT("Request Type", ': ', "Request Details") AS description, 
  "Media URL" AS description_details,
  Category AS category,
  "Request Details" AS activity,
  "Status" AS status,
  Source AS method,
  "Status Notes" AS outcome,
  LOWER("Police District") AS neighborhoods, 
FROM san_francisco
),

BerkeleyNormalized AS 
(
SELECT
  'Berkeley' AS city,
  'CA' AS state, 
  Date_Opened  AS start_event_date,
  STRPTIME(Date_Closed, '%m/%d/%Y %I:%M:%S %p') AS end_event_date,
  Longitude AS longitude,
  Latitude AS latitude,
  Case_ID AS ticket_id,
  LOWER(Street_Address) AS street_address,
  array_slice(Street_Address, LENGTH(SPLIT(Street_Address, ' ')[1]) + 2, 100) AS street,
  NULL AS zipcode,
  SPLIT(REPLACE(REPLACE(Request_Detail, 'Commercial ', ''), 'Residential ', ''), ' - ')[1] AS description, 
  REPLACE(REPLACE(Request_Detail, 'Commercial ', ''), 'Residential ', '') AS description_details, 
  Request_Category AS category,
  Request_SubCategory AS activity,
  Case_Status AS status,
  NULL AS method,
  NULL AS outcome,
  LOWER(Neighborhood) AS neighborhood,
FROM berkeley
),

SanJoseNormalized AS 
(
SELECT
  'San Jose' AS city,
  'CA' AS state,
  "Date Created" AS start_event_date,
  "Date Last Updated" AS end_event_date,
  Longitude AS longitude,
  Latitude AS latitude,
  Incident_ID AS ticket_id,
  NULL AS street_address,
  NULL AS street,
  NULL AS zipcode,
  Category AS description, 
  NULL  AS description_details,
  "Service Type"  AS category,
  Department AS activity,
  Status AS status,
  Source AS method,
  null AS outcome,
  'san jose'  AS neighborhood,
FROM sanjose
),

BostonNormalized AS 
(
SELECT
  'Boston' AS city,
  'MA' AS state,
  open_dt AS start_event_date,
  closed_dt AS end_event_date,
  longitude AS longitude,
  latitude AS latitude,
  case_enquiry_id AS ticket_id,
  location_street_name AS street_address,
  LOWER(array_slice(location_street_name, LENGTH(SPLIT(location_street_name, ' ')[1]) + 2, 100)) AS street,
  location_zipcode AS zipcode,
  reason AS description, 
  NULL AS description_details,
  subject AS category,
  type AS activity,
  case_status AS status,
  source AS method,
  REPLACE(closure_reason, 'Case Closed ', '') AS outcome,
  LOWER(City) AS neighborhood,
FROM boston
),

AustinNormalized AS 
(
SELECT
  'Austin' AS city,
  'TX' AS state,
  "Created Date" AS start_event_date,
  "Last Update Date" AS end_event_date,
  "Longitude Coordinate" AS longitude,
  "Latitude Coordinate" AS latitude,
  "Service Request (SR) Number" AS ticket_id,
  CONCAT("Street Number", ' ', "Street Name") AS street_address,
  LOWER("Street Name") AS street,
  "Zip Code" AS zipcode,
  "SR Description" AS description, 
  "SR Description" AS description_details,
  SPLIT("SR Description", ' - ')[1] AS category,
  NULL AS activity,
  "SR Status" AS status,
  "Method Received" AS method,
  NULL AS outcome,
  LOWER(City) AS neighborhood,
FROM austin
),

Together AS (
SELECT * FROM BerkeleyNormalized
  UNION ALL
SELECT * FROM SanJoseNormalized
  UNION ALL 
SELECT * FROM BostonNormalized
  UNION ALL 
SELECT * FROM AustinNormalized
  UNION ALL
SELECT * FROM SanFranciscoNormalized
)


SELECT
  DATE_DIFF('HOUR', start_event_date, end_event_date) AS date_diff_in_hours,
  CASE 
    WHEN LOWER(status) IN ('open', 'new', 'in progress') THEN 'Active' 
    WHEN status IS NULL THEN 'Unknown' 
    ELSE 'Closed' 
    END AS status_type,
  CAST(ticket_id AS VARCHAR) AS ticket_id,
  LOWER(category) AS category,
  * exclude(ticket_id, category)
FROM Together 
WHERE start_event_date >= '2023-01-01' 
