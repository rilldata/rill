WITH 
    NormalizedData AS (
        SELECT 
            YEAR(strptime("datetime", '%-m/%-d/%Y %-H:%M')) AS event_year,
            YEAR("date posted") AS report_year, 
            YEAR(strptime("datetime", '%-m/%-d/%Y %-H:%M')) = YEAR("date posted") AS report_same_year, 
            "duration (seconds)" AS duration_in_seconds, 
            city,
            UCASE(state) AS state_code,
            UCASE(country) AS country_code,
            shape AS shape_of_ufo,
            LCASE(comments) AS comments,
            CASE 
                WHEN lcase(comments) LIKE '%fast%' OR lcase(comments) LIKE '%quick%' THEN 'fast'
                WHEN lcase(comments) LIKE '%slow%'  OR lcase(comments) LIKE '%steady%' THEN 'slow'
                ELSE 'Unknown' END
            AS speed_info, 
            CASE 
                WHEN lcase(comments) LIKE '%big%' OR lcase(comments) LIKE '%large%' THEN 'big'
                WHEN lcase(comments) LIKE '%small%' OR lcase(comments) LIKE '%little%'  THEN 'small'
                ELSE 'Unknown' END
            AS size_info
        FROM UFO_Reports
        WHERE "duration (seconds)" < 1000
    )
    SELECT 
        *
        FROM NormalizedData
        WHERE country_code = 'US' AND report_same_year