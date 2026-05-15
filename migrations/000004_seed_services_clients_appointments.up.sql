-- ============================================================
-- master_settings: дефолтные настройки слотов для сид-мастеров
-- ============================================================
INSERT INTO master_settings (master_id, slot_step_minutes, lead_time_minutes)
SELECT
    m.id,
    CASE (ROW_NUMBER() OVER (ORDER BY m.created_at, m.id))::INT % 3
        WHEN 0 THEN 30
        WHEN 1 THEN 60
        ELSE 15
    END,
    CASE (ROW_NUMBER() OVER (ORDER BY m.created_at, m.id))::INT % 3
        WHEN 0 THEN 0
        WHEN 1 THEN 60
        ELSE 30
    END
FROM masters m
JOIN "user" u ON u.user_id = m.user_id
WHERE u.email LIKE '%@seed.okoshki.ru';

-- ============================================================
-- master_work_intervals: 09:00–20:00 каждый день в окне 2026-04-30..2026-06-02
-- (-15..+18 от сегодняшней даты 2026-05-15)
-- ============================================================
INSERT INTO master_work_intervals (master_id, work_date, start_time, end_time)
SELECT m.id, d::DATE, TIME '09:00', TIME '20:00'
FROM masters m
JOIN "user" u ON u.user_id = m.user_id
CROSS JOIN generate_series(DATE '2026-04-30', DATE '2026-06-02', INTERVAL '1 day') d
WHERE u.email LIKE '%@seed.okoshki.ru';

-- ============================================================
-- Клиенты, услуги мастеров и записи
-- ============================================================
DO $$
DECLARE
    today_date DATE := DATE '2026-05-15';
    pwd_hash   TEXT := '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi';

    client_firsts TEXT[] := ARRAY[
        'Анна','Мария','Елена','Ольга','Наталья',
        'Татьяна','Ирина','Светлана','Юлия','Екатерина',
        'Алина','Виктория','Дарья','Ксения','Полина',
        'Валерия','Александр','Дмитрий','Андрей','Сергей',
        'Алексей','Михаил','Иван','Николай','Артём'
    ];
    client_lasts TEXT[] := ARRAY[
        'Иванова','Смирнова','Кузнецова','Попова','Соколова',
        'Лебедева','Козлова','Новикова','Морозова','Петрова',
        'Волкова','Соловьёва','Васильева','Зайцева','Павлова',
        'Семёнова','Иванов','Смирнов','Кузнецов','Попов',
        'Соколов','Лебедев','Козлов','Новиков','Морозов'
    ];

    -- Каталог услуг: 9 категорий × 5 услуг.
    -- Индекс в массиве = (cat_idx - 1) * 5 + svc_local_idx (оба 1-based).
    svc_titles TEXT[] := ARRAY[
        'Маникюр','Педикюр','Покрытие гель-лаком','Наращивание ногтей','Дизайн ногтей',
        'Женская стрижка','Окрашивание','Укладка','Уход за волосами','Кератиновое выпрямление',
        'Мужская стрижка','Стрижка бороды','Бритьё опасной бритвой','Камуфляж седины','Детская стрижка',
        'Архитектура бровей','Ламинирование бровей','Окрашивание бровей','Наращивание ресниц','Ламинирование ресниц',
        'Дневной макияж','Вечерний макияж','Свадебный макияж','Возрастной макияж','Макияж для фотосессии',
        'Чистка лица','Пилинг','Мезотерапия','Биоревитализация','Контурная пластика',
        'Классический массаж','Антицеллюлитный массаж','Лимфодренажный массаж','Спортивный массаж','SPA-программа',
        'Шугаринг ног','Шугаринг рук','Восковая эпиляция','Эпиляция зоны бикини','Эпиляция подмышек',
        'Тату маленькое','Тату среднее','Перманент бровей','Перманент губ','Перманент стрелок'
    ];
    svc_durations INT[] := ARRAY[
         60, 75, 90,120, 45,
         60,180, 45, 60,180,
         45, 30, 45, 60, 30,
         45, 60, 30,150, 90,
         60, 90,120, 75, 90,
         90, 60, 45, 60, 45,
         60, 90, 75, 60,120,
         60, 30, 75, 45, 30,
         60,180,120,180, 90
    ];
    svc_prices BIGINT[] := ARRAY[
        1500,2000,2500,3500,1000,
        2000,5000,1500,2500,8000,
        1500,1000,1200,2000,1000,
        1500,2000,1000,3000,2500,
        2500,3500,6000,4000,4500,
        3500,4000,5500,8000,12000,
        2500,3500,3000,3000,5000,
        2000,1000,2500,2000,800,
        5000,15000,12000,15000,10000
    ];
    svc_descriptions TEXT[] := ARRAY[
        'Классический маникюр с обработкой кутикулы',
        'Аппаратный педикюр с обработкой стоп',
        'Долговременное покрытие гель-лаком до 3 недель',
        'Наращивание ногтей акрилом или гелем',
        'Авторский дизайн ногтей любой сложности',
        'Стрижка с консультацией стилиста',
        'Окрашивание профессиональными красителями',
        'Укладка феном или плойкой',
        'Восстанавливающий уход за повреждёнными волосами',
        'Кератиновое выпрямление и разглаживание волос',
        'Мужская стрижка ножницами и машинкой',
        'Моделирование и оформление бороды',
        'Бритьё опасной бритвой с горячими компрессами',
        'Тонирование для маскировки седых волос',
        'Детская стрижка от 3 до 12 лет',
        'Коррекция формы бровей пинцетом и воском',
        'Долговременная укладка бровей',
        'Окрашивание бровей хной или краской',
        'Наращивание ресниц: классика, 2D, 3D',
        'Ламинирование и ботокс ресниц',
        'Дневной макияж для повседневной носки',
        'Вечерний макияж со стойкой фиксацией',
        'Свадебный макияж с пробным образом',
        'Лифтинг-макияж и возрастной мейк-ап',
        'Макияж для фото- и видеосъёмки',
        'Комбинированная чистка лица',
        'Химический пилинг лица',
        'Инъекционная мезотерапия',
        'Биоревитализация препаратами гиалуроновой кислоты',
        'Контурная пластика губ и носогубных складок',
        'Общеоздоровительный классический массаж',
        'Антицеллюлитный массаж проблемных зон',
        'Лимфодренажный массаж тела',
        'Спортивный массаж для восстановления',
        'Комплексная SPA-программа с обёртыванием',
        'Шугаринг ног до колена или полностью',
        'Шугаринг рук до локтя или полностью',
        'Восковая депиляция любых зон',
        'Эпиляция зоны бикини: классика или глубокое',
        'Эпиляция зоны подмышечных впадин',
        'Маленькая татуировка до 5 см',
        'Татуировка среднего размера до 15 см',
        'Перманентный макияж бровей',
        'Перманентный макияж губ',
        'Перманентный макияж стрелок'
    ];

    statuses TEXT[] := ARRAY['completed','cancelled','rejected','confirmed','pending'];

    nail_id UUID; hair_id UUID; barber_id UUID;
    brows_id UUID; makeup_id UUID; cosmo_id UUID;
    massage_id UUID; epil_id UUID; tattoo_id UUID;

    client_ids UUID[] := ARRAY[]::UUID[];

    new_uid UUID;
    i INT;
    j INT;

    master_rec RECORD;
    master_idx INT := 0;
    master_cat_idx INT;

    n_services INT;
    svc_local_idx INT;
    svc_catalog_idx INT;
    new_service_id UUID;
    master_service_ids UUID[];
    master_service_durations INT[];

    appt_idx INT;
    appt_day_offset INT;
    appt_start TIMESTAMPTZ;
    appt_status TEXT;
    chosen_svc_local_idx INT;
    chosen_svc_id UUID;
    chosen_duration INT;
    chosen_client_id UUID;

BEGIN
    -- 1) Маппинг category_id -> номер категории (1..9)
    SELECT id INTO nail_id    FROM category WHERE name = 'Ногтевой сервис';
    SELECT id INTO hair_id    FROM category WHERE name = 'Волосы';
    SELECT id INTO barber_id  FROM category WHERE name = 'Барбершоп';
    SELECT id INTO brows_id   FROM category WHERE name = 'Брови и ресницы';
    SELECT id INTO makeup_id  FROM category WHERE name = 'Макияж';
    SELECT id INTO cosmo_id   FROM category WHERE name = 'Косметология';
    SELECT id INTO massage_id FROM category WHERE name = 'Массаж и SPA';
    SELECT id INTO epil_id    FROM category WHERE name = 'Эпиляция и депиляция';
    SELECT id INTO tattoo_id  FROM category WHERE name = 'Тату и перманент';

    -- 2) 50 клиентов
    FOR i IN 1..50 LOOP
        new_uid := uuid_generate_v4();

        INSERT INTO "user" (user_id, email, password_hash, role)
        VALUES (new_uid, 'client_' || i || '@seed.okoshki.ru', pwd_hash, 'client');

        INSERT INTO clients (user_id, first_name, last_name, phone)
        VALUES (
            new_uid,
            client_firsts[((i - 1) % 25) + 1],
            client_lasts[((i - 1) % 25) + 1],
            '+78' || LPAD((100000000 + i * 137)::TEXT, 9, '0')
        );

        client_ids := array_append(client_ids, new_uid);
    END LOOP;

    -- 3) Для каждого мастера: услуги (2..5) + 5 записей (по одной на статус)
    FOR master_rec IN
        SELECT m.id, m.category_id, m.city, m.address, m.lat, m.lon
        FROM masters m
        JOIN "user" u ON u.user_id = m.user_id
        WHERE u.email LIKE '%@seed.okoshki.ru'
        ORDER BY m.created_at, m.id
    LOOP
        master_idx := master_idx + 1;

        master_cat_idx := CASE master_rec.category_id
            WHEN nail_id    THEN 1
            WHEN hair_id    THEN 2
            WHEN barber_id  THEN 3
            WHEN brows_id   THEN 4
            WHEN makeup_id  THEN 5
            WHEN cosmo_id   THEN 6
            WHEN massage_id THEN 7
            WHEN epil_id    THEN 8
            WHEN tattoo_id  THEN 9
        END;

        -- 2..5 услуг, варьируется по master_idx
        n_services := 2 + (master_idx % 4);
        master_service_ids := ARRAY[]::UUID[];
        master_service_durations := ARRAY[]::INT[];

        FOR j IN 1..n_services LOOP
            -- циклически берём услуги из подкаталога категории со сдвигом по master_idx
            svc_local_idx   := ((master_idx + j - 2) % 5) + 1;
            svc_catalog_idx := (master_cat_idx - 1) * 5 + svc_local_idx;

            INSERT INTO master_services (
                master_id, category_id, title, address, city, description,
                price, duration_minutes, lat, lon
            ) VALUES (
                master_rec.id,
                master_rec.category_id,
                svc_titles[svc_catalog_idx],
                master_rec.address,
                master_rec.city,
                svc_descriptions[svc_catalog_idx],
                svc_prices[svc_catalog_idx],
                svc_durations[svc_catalog_idx],
                master_rec.lat,
                master_rec.lon
            ) RETURNING id INTO new_service_id;

            master_service_ids       := array_append(master_service_ids, new_service_id);
            master_service_durations := array_append(master_service_durations, svc_durations[svc_catalog_idx]);
        END LOOP;

        -- 5 записей: completed, cancelled, rejected (прошлое) + confirmed, pending (будущее)
        FOR appt_idx IN 1..5 LOOP
            appt_status := statuses[appt_idx];

            CASE appt_status
                WHEN 'completed' THEN
                    appt_day_offset := -15 + (master_idx % 6);
                    appt_start := (today_date + appt_day_offset)::TIMESTAMPTZ + INTERVAL '10 hours';
                WHEN 'cancelled' THEN
                    appt_day_offset := -9 + (master_idx % 5);
                    appt_start := (today_date + appt_day_offset)::TIMESTAMPTZ + INTERVAL '12 hours';
                WHEN 'rejected' THEN
                    appt_day_offset := -4 + (master_idx % 4);
                    appt_start := (today_date + appt_day_offset)::TIMESTAMPTZ + INTERVAL '14 hours 30 minutes';
                WHEN 'confirmed' THEN
                    appt_day_offset := 2 + (master_idx % 7);
                    appt_start := (today_date + appt_day_offset)::TIMESTAMPTZ + INTERVAL '11 hours';
                WHEN 'pending' THEN
                    appt_day_offset := 9 + (master_idx % 9);
                    appt_start := (today_date + appt_day_offset)::TIMESTAMPTZ + INTERVAL '15 hours';
            END CASE;

            chosen_svc_local_idx := ((master_idx + appt_idx - 1) % n_services) + 1;
            chosen_svc_id        := master_service_ids[chosen_svc_local_idx];
            chosen_duration      := master_service_durations[chosen_svc_local_idx];

            chosen_client_id := client_ids[((master_idx * 5 + appt_idx - 1) % 50) + 1];

            INSERT INTO appointments (
                client_id, master_id, service_id, start_at, end_at, status, client_comment
            ) VALUES (
                chosen_client_id,
                master_rec.id,
                chosen_svc_id,
                appt_start,
                appt_start + (chosen_duration * INTERVAL '1 minute'),
                appt_status,
                CASE appt_idx % 3
                    WHEN 0 THEN 'Пожалуйста, без сильных запахов.'
                    WHEN 1 THEN NULL
                    ELSE        'Буду немного раньше, не страшно?'
                END
            );
        END LOOP;
    END LOOP;
END $$;
