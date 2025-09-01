-- We stored HTML URL earlier, but now we are using clone URL
UPDATE projects 
SET github_url = concat(github_url, '.git') 
WHERE github_url IS NOT NULL 
  AND github_url NOT LIKE '%.git';
