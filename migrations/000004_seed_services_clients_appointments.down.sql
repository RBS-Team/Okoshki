-- Удаляем записи, услуги, расписание и настройки сид-мастеров
DELETE FROM appointments
WHERE master_id IN (
    SELECT m.id FROM masters m
    JOIN "user" u ON u.user_id = m.user_id
    WHERE u.email LIKE '%@seed.okoshki.ru'
);

DELETE FROM master_services
WHERE master_id IN (
    SELECT m.id FROM masters m
    JOIN "user" u ON u.user_id = m.user_id
    WHERE u.email LIKE '%@seed.okoshki.ru'
);

DELETE FROM master_work_intervals
WHERE master_id IN (
    SELECT m.id FROM masters m
    JOIN "user" u ON u.user_id = m.user_id
    WHERE u.email LIKE '%@seed.okoshki.ru'
);

DELETE FROM master_settings
WHERE master_id IN (
    SELECT m.id FROM masters m
    JOIN "user" u ON u.user_id = m.user_id
    WHERE u.email LIKE '%@seed.okoshki.ru'
);

-- Клиентов сидов сносим вместе с user (clients.user_id ON DELETE CASCADE,
-- appointments по client_id уже удалены выше)
DELETE FROM "user" WHERE email LIKE 'client_%@seed.okoshki.ru';
