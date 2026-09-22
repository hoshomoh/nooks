-- Every Instance that exists keeps the door it already had.
--
-- instance.public_signup was stored, drawn in Settings as a toggle, and read by nothing:
-- RequestJoin always created a request whatever it said, so an Admin who turned it off
-- was told something happened and nothing did. It now decides whether a stranger may ask
-- for an account at all.
--
-- Turning it on here rather than honouring what is stored, because what is stored was
-- never acted on. An Instance running today accepts requests whatever its toggle reads,
-- so writing false through would close a door that is open and take a working sign-in
-- link away from a household that never chose to.
--
-- New Instances are opened the same way, by CompleteSetup, so the two agree.
UPDATE setting SET value = 'true' WHERE key = 'instance.public_signup';
INSERT INTO setting (key, value)
SELECT 'instance.public_signup', 'true'
WHERE NOT EXISTS (SELECT 1 FROM setting WHERE key = 'instance.public_signup');
