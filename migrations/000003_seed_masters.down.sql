DELETE FROM masters
WHERE user_id IN (
    SELECT user_id FROM "user" WHERE email LIKE '%@seed.okoshki.ru'
);

DELETE FROM "user" WHERE email LIKE '%@seed.okoshki.ru';
