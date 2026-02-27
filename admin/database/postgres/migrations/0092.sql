UPDATE deployments 
SET desired_status = CASE 
    WHEN status IN (1, 2, 4, 6) THEN 2  -- Pending, Running, Errored, Updating -> Running
    WHEN status IN (5, 7) THEN 5        -- Stopped, Stopping -> Stopped
    WHEN status IN (8, 9) THEN 9        -- Deleting, Deleted -> Deleted
    ELSE 2                              -- Default to Running for any other case
END
WHERE desired_status = 0;
