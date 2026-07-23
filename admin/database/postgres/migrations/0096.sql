-- User codes are lookup keys while a device request is pending. Keeping them
-- unique prevents a confirmation from approving an arbitrary colliding request.
CREATE UNIQUE INDEX device_auth_codes_pending_user_code_idx
ON device_auth_codes (user_code)
WHERE approval_state = 0;
