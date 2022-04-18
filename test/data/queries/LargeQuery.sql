-- Found on Quora. Cannot find the exact source.
SELECT
       Max("NEW126"."Parent1") AS "Parent1", Max("NEW126"."Parent2") AS "Parent2",
       Max("NEW126"."Child1") AS "Child1", Max("NEW126"."Child2") AS "Child2", Max("NEW126"."Child3") AS "Child3",
       Max("NEW126"."Child4") AS "Child4", Max("NEW126"."Child5") AS "Child5", Max("NEW126"."Child6") AS "Child6",
       SEC_TO_TIME(SUM("NEW126"."Child7")) AS "Value1",
       CASE WHEN "NEW126"."Parent1" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Parent1" AS VARCHAR) END
           || '^^' || CASE WHEN "NEW126"."Parent2" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Parent2" AS VARCHAR) END AS "ParentGroup",
       CASE WHEN "NEW126"."Child1" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child1" AS VARCHAR) END
           || '^^' || CASE WHEN "NEW126"."Child2" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child2" AS VARCHAR) END
           || '^^' || CASE WHEN "NEW126"."Child3" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child3" AS VARCHAR) END
           || '^^' || CASE WHEN "NEW126"."Child4" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child4" AS VARCHAR) END
           || '^^' || CASE WHEN "NEW126"."Child5" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child5" AS VARCHAR) END
           || '^^' || CASE WHEN "NEW126"."Child6" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child6" AS VARCHAR) END AS "ChildGroup"
FROM
     (
         SELECT
                Max("Vehicles"."Vehicle") AS "Parent1",
                Max(Date(Add_Hours("GeoTabTrips"."NextTripStartTime",1))) AS "Parent2",
                Max("GeoTabTrips"."GeoTabTripID") AS "Child1",
                Max(CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) > 12 THEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) - 12
                         ELSE HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) END ||':'||
                    CASE WHEN MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) < 10 THEN '0' || MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime")))
                         ELSE TO_VARCHAR(MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime")))) END ||
                    CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) > 12 THEN ' PM' ELSE ' AM' END) AS "Child2",
                Max(CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) > 12 THEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) - 12
                         ELSE HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) END ||':'||
                    CASE WHEN MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) < 10 THEN '0' || MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime")))
                         ELSE TO_VARCHAR(MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime")))) END||
                    CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) > 12 THEN ' PM' ELSE ' AM' END) AS "Child3",
                Max(CASE WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '3360 Chelsea Rd%' THEN 'Dahlheimer-Warehouse' WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '923 Wright St%' THEN 'D &amp;amp; D Beverage' ELSE "Customers"."Company" END) AS "Child4",
                Max(IFNULL("Customers"."Address","GeoTabTrips"."Address")) AS "Child5",
                Max(CASE WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '3360 Chelsea Rd%' THEN 'Monticello' WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '923 Wright St%' THEN 'Brainerd' ELSE "Customers"."City" END) AS "Child6",
                Max(CAST(CASE WHEN (DATE("GeoTabTrips"."StopTime")!=DATE("GeoTabTrips"."NextTripStartTime")) THEN ("GeoTabTrips"."StopDuration"-"GeoTabTrips"."StopDuration") ELSE ("GeoTabTrips"."StopDuration") END AS INT)) AS "Child7"
         FROM
              "Vehicles" INNER JOIN "GPSDevices" ON "Vehicles"."GPSDeviceID" = "GPSDevices"."GPSDeviceID"
                  INNER JOIN "GeoTabTrips" ON "GPSDevices"."GPSDeviceID" = "GeoTabTrips"."GPSDeviceID"
                  LEFT JOIN (
                      SELECT
                             Max("GeoTabTrips"."GeoTabTripID") AS "Parent1",
                             Min(CAST(CASE WHEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude") < (.09)) THEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude")*1093.613) END AS DECIMAL(18,6))) AS "Value1"
                      FROM
                           "Vehicles" INNER JOIN "GPSDevices" ON "Vehicles"."GPSDeviceID" = "GPSDevices"."GPSDeviceID"
                               INNER JOIN "GeoTabTrips" ON "GPSDevices"."GPSDeviceID" = "GeoTabTrips"."GPSDeviceID"
                               INNER JOIN "Customers" ON "Vehicles"."LocationID" = "Customers"."LocationID"
                      WHERE "GeoTabTrips"."StopDuration" < 240 AND
                            (("GeoTabTrips"."StartTime" > '2019-01-21 23:59:59' AND
                              "GeoTabTrips"."StartTime" < '2019-01-28 23:59:59')) AND
                            "Customers"."CustomerTypeID" Not IN ( 49 , 48 , 25 , 23 , 42 , 10 , 37 ) AND
                            ( "Customers"."AccountStatus" = 'Active' )
                      GROUP BY CASE WHEN "GeoTabTrips"."GeoTabTripID" IS NULL THEN 'Is Null' ELSE CAST("GeoTabTrips"."GeoTabTripID" AS VARCHAR) END
                  ) AS "MinCustomerDistance" ON "GeoTabTrips"."GeoTabTripID" = "MinCustomerDistance"."Parent1"
                  LEFT JOIN (
                      SELECT
                             Max("GeoTabTrips"."GeoTabTripID") AS "Parent1",
                             Max(CAST(CASE WHEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude") < (.09)) THEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude")*1093.613) END AS DECIMAL(18,6))) AS "Parent2",
                             Max(CAST(CASE WHEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude") < (.09)) THEN "Customers"."CustomerID" END AS INT)) AS "Parent3"
                      FROM "Vehicles" INNER JOIN "GPSDevices" ON "Vehicles"."GPSDeviceID" = "GPSDevices"."GPSDeviceID"
                          INNER JOIN "GeoTabTrips" ON "GPSDevices"."GPSDeviceID" = "GeoTabTrips"."GPSDeviceID"
                          INNER JOIN "Customers" ON "Vehicles"."LocationID" = "Customers"."LocationID"
                      WHERE "GeoTabTrips"."StopDuration" > 240 AND (("GeoTabTrips"."StartTime" > '2019-01-21 23:59:59' AND "GeoTabTrips"."StartTime" < '2019-01-28 23:59:59')) AND "Customers"."CustomerTypeID" Not IN ( 49 , 48 , 25 , 23 , 42 , 10 , 37 ) AND ( "Customers"."AccountStatus" = 'Active' )
                      GROUP BY CASE WHEN "GeoTabTrips"."GeoTabTripID" IS NULL THEN 'Is Null' ELSE CAST("GeoTabTrips"."GeoTabTripID" AS VARCHAR) END || '^^' ||
                               CASE WHEN CAST(CASE WHEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude")<(.09)) THEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude")*1093.613) END AS DECIMAL(18,6)) IS NULL THEN 'Is Null' ELSE CAST(CAST(
                                   CASE WHEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude")<(.09)) THEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude")*1093.613) END AS DECIMAL(18,6)) AS VARCHAR) END || '^^' ||
                                   CASE WHEN CAST(CASE WHEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude") < (.09)) THEN "Customers"."CustomerID" END AS INT) IS NULL THEN 'Is Null' ELSE CAST(CAST(CASE WHEN (HAVERSINE("GeoTabTrips"."StopLongitude","GeoTabTrips"."StopLatitude","Customers"."Longitude","Customers"."Latitude")<(.09)) THEN "Customers"."CustomerID" END AS INT) AS VARCHAR) END
                  ) AS "CustomerwithminDist" ON "GeoTabTrips"."GeoTabTripID" = "CustomerwithminDist"."Parent1" AND "CustomerwithminDist"."Parent2" = "MinCustomerDistance"."Value1"
                  LEFT JOIN "Customers" ON "Customers"."CustomerID" = IFNULL("GeoTabTrips"."CustomerID","CustomerwithminDist"."Parent3") WHERE "GeoTabTrips"."StopDuration" > 240 AND (("GeoTabTrips"."NextTripStartTime" >= '2019-01-25 00:00:00' AND "GeoTabTrips"."NextTripStartTime" < '2019-01-25 23:59:59'))
         GROUP BY (CASE WHEN "Vehicles"."Vehicle" IS NULL THEN 'Is Null' ELSE CAST("Vehicles"."Vehicle" AS VARCHAR) END || '^^' ||
                   CASE WHEN Date(Add_Hours("GeoTabTrips"."NextTripStartTime",1)) IS NULL THEN 'Is Null' ELSE CAST(Date(Add_Hours("GeoTabTrips"."NextTripStartTime",1)) AS VARCHAR) END,
                   CASE WHEN "GeoTabTrips"."GeoTabTripID" IS NULL THEN 'Is Null' ELSE CAST("GeoTabTrips"."GeoTabTripID" AS VARCHAR) END || '^^' ||
                   CASE WHEN CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) > 12 THEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) - 12 ELSE HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) END ||':'||
                             CASE WHEN MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) < 10 THEN '0' || MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) ELSE TO_VARCHAR(MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime")))) END||
                             CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) >=12 THEN ' PM' ELSE ' AM' END IS NULL THEN 'Is Null' ELSE CAST(
                                 CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) > 12 THEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) - 12 ELSE HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) END ||':'||
                                 CASE WHEN MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) < 10 THEN '0' || MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) ELSE TO_VARCHAR(MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."StopTime")))) END||
                                 CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."StopTime"))) >=12 THEN ' PM' ELSE ' AM' END AS VARCHAR) END || '^^' ||
                   CASE WHEN CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) > 12 THEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) - 12 ELSE HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) END ||':'||
                             CASE WHEN MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) < 10 THEN '0' || MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) ELSE TO_VARCHAR(MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime")))) END||
                             CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) >=12 THEN ' PM' ELSE ' AM' END IS NULL THEN 'Is Null' ELSE CAST(
                                 CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) >12 THEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) - 12 ELSE HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) END ||':'||
                                 CASE WHEN MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) < 10 THEN '0' || MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) ELSE TO_VARCHAR(MINUTE(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime")))) END||
                                 CASE WHEN HOUR(DATEADD( hour, 1, ("GeoTabTrips"."NextTripStartTime"))) >=12 THEN ' PM' ELSE ' AM' END AS VARCHAR) END || '^^' ||
                   CASE WHEN CASE WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '3360 Chelsea Rd%' THEN 'Dahlheimer-Warehouse' WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '923 Wright St%' THEN 'D &amp;amp; D Beverage' ELSE "Customers"."Company" END IS NULL THEN 'Is Null' ELSE CAST(
                       CASE WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '3360 Chelsea Rd%' THEN 'Dahlheimer-Warehouse' WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '923 Wright St%' THEN 'D &amp;amp; D Beverage' ELSE "Customers"."Company" END AS VARCHAR) END || '^^' ||
                   CASE WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") IS NULL THEN 'Is Null' ELSE CAST(IFNULL("Customers"."Address","GeoTabTrips"."Address") AS VARCHAR) END || '^^' ||
                   CASE WHEN CASE WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '3360 Chelsea Rd%' THEN 'Monticello' WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '923 Wright St%' THEN 'Brainerd' ELSE "Customers"."City" END IS NULL THEN 'Is Null' ELSE CAST(
                       CASE WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '3360 Chelsea Rd%' THEN 'Monticello' WHEN IFNULL("Customers"."Address","GeoTabTrips"."Address") LIKE '923 Wright St%' THEN 'Brainerd' ELSE "Customers"."City" END AS VARCHAR) END || '^^' ||
                   CASE WHEN CAST(CASE WHEN (DATE("GeoTabTrips"."StopTime")!=DATE("GeoTabTrips"."NextTripStartTime")) THEN ("GeoTabTrips"."StopDuration"-"GeoTabTrips"."StopDuration") ELSE ("GeoTabTrips"."StopDuration") END AS INT) IS NULL THEN 'Is Null' ELSE CAST(
                       CAST(CASE WHEN (DATE("GeoTabTrips"."StopTime")!=DATE("GeoTabTrips"."NextTripStartTime")) THEN ("GeoTabTrips"."StopDuration"-"GeoTabTrips"."StopDuration") ELSE ("GeoTabTrips"."StopDuration") END AS INT) AS VARCHAR) END)
     ) AS "NEW126"
GROUP BY ROLLUP(
  CASE WHEN "NEW126"."Parent1" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Parent1" AS VARCHAR) END || '^^' ||
    CASE WHEN "NEW126"."Parent2" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Parent2" AS VARCHAR) END,
  CASE WHEN "NEW126"."Child1" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child1" AS VARCHAR) END || '^^' ||
    CASE WHEN "NEW126"."Child2" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child2" AS VARCHAR) END || '^^' ||
    CASE WHEN "NEW126"."Child3" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child3" AS VARCHAR) END || '^^' ||
    CASE WHEN "NEW126"."Child4" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child4" AS VARCHAR) END || '^^' ||
    CASE WHEN "NEW126"."Child5" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child5" AS VARCHAR) END || '^^' ||
    CASE WHEN "NEW126"."Child6" IS NULL THEN 'Is Null' ELSE CAST("NEW126"."Child6" AS VARCHAR) END)
ORDER BY CASE WHEN "ParentGroup" IS NULL AND "ChildGroup" IS NULL THEN 0 WHEN "ParentGroup" IS NULL AND "ChildGroup" IS NULL THEN 1 ELSE 2 END , "Parent1" , "Parent2" , "Child1" , "Child2" , "Child3" , "Child4" , "Child5" , "Child6" ;
