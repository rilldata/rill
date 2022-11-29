WITH 

TransformData AS
(
SELECT 
* EXCLUDE (
  event_time, 
  last_edited_date, 
  event_title,
  event_description, 
  location_description, 
  notes,
  source_link,
  photo_link, 
  storm_name,
  country_code,
  event_import_source,
  submitted_date,
  created_date
  )
FROM global_landslide_catalog

)

SELECT 
  *
FROM 
  TransformData

