DO $$
DECLARE
    female_firsts TEXT[] := ARRAY[
        'Анна','Мария','Елена','Ольга','Наталья','Татьяна','Ирина','Светлана','Людмила','Юлия',
        'Екатерина','Алина','Виктория','Дарья','Ксения','Полина','Валерия','Надежда','Галина','Вера',
        'Зоя','Тамара','Лариса','Нина','Марина'
    ];
    female_lasts TEXT[] := ARRAY[
        'Иванова','Смирнова','Кузнецова','Попова','Соколова','Лебедева','Козлова','Новикова','Морозова','Петрова',
        'Волкова','Соловьёва','Васильева','Зайцева','Павлова','Семёнова','Голубева','Виноградова','Богданова','Воробьёва',
        'Фёдорова','Михайлова','Беляева','Тарасова','Белова'
    ];
    male_firsts TEXT[] := ARRAY[
        'Александр','Дмитрий','Андрей','Сергей','Алексей','Михаил','Иван','Николай','Владимир','Артём',
        'Роман','Максим','Евгений','Павел','Денис','Илья','Константин','Игорь','Виктор','Антон',
        'Вадим','Леонид','Олег','Станислав','Тимур'
    ];
    male_lasts TEXT[] := ARRAY[
        'Иванов','Смирнов','Кузнецов','Попов','Соколов','Лебедев','Козлов','Новиков','Морозов','Петров',
        'Волков','Соловьёв','Васильев','Зайцев','Павлов','Семёнов','Голубев','Виноградов','Богданов','Воробьёв',
        'Фёдоров','Михайлов','Беляев','Тарасов','Белов'
    ];
    cities     TEXT[]   := ARRAY[
        'Москва','Санкт-Петербург','Новосибирск','Екатеринбург','Казань',
        'Нижний Новгород','Челябинск','Самара','Омск','Ростов-на-Дону',
        'Уфа','Красноярск','Пермь','Воронеж','Волгоград'
    ];
    city_lats  FLOAT8[] := ARRAY[
        55.7558,59.9311,54.9893,56.8389,55.7887,
        56.2965,55.1644,53.2038,54.9885,47.2357,
        54.7388,56.0153,58.0105,51.6755,48.7194
    ];
    city_lons  FLOAT8[] := ARRAY[
        37.6173,30.3609,82.9182,60.6057,49.1221,
        43.9361,61.4368,50.1606,73.3242,39.7015,
        55.9721,92.8932,56.2502,39.2088,44.5018
    ];
    streets TEXT[] := ARRAY[
        'ул. Ленина','ул. Пушкина','проспект Мира','ул. Гагарина','ул. Советская',
        'ул. Центральная','ул. Садовая','проспект Победы','ул. Молодёжная','ул. Комсомольская',
        'ул. Новая','ул. Школьная','ул. Лесная','ул. Озёрная','ул. Строителей'
    ];

    nail_id UUID; hair_id UUID; barber_id UUID;
    brows_id UUID; makeup_id UUID; cosmo_id UUID;
    massage_id UUID; epil_id UUID; tattoo_id UUID;

    cat_ids   UUID[];
    cat_slugs TEXT[];

    i         INT;
    cat_idx   INT;
    cat_id    UUID;
    cat_slug  TEXT;
    new_uid   UUID;
    fname     TEXT;
    lname     TEXT;
    city_i    INT;
    street_i  INT;
    bio_text  TEXT;
    rating_v  DECIMAL(3,2);
    review_v  INT;
    is_female BOOLEAN;
    exp_years INT;

    pwd_hash TEXT := '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi';

BEGIN
    SELECT id INTO nail_id    FROM category WHERE name = 'Ногтевой сервис';
    SELECT id INTO hair_id    FROM category WHERE name = 'Волосы';
    SELECT id INTO barber_id  FROM category WHERE name = 'Барбершоп';
    SELECT id INTO brows_id   FROM category WHERE name = 'Брови и ресницы';
    SELECT id INTO makeup_id  FROM category WHERE name = 'Макияж';
    SELECT id INTO cosmo_id   FROM category WHERE name = 'Косметология';
    SELECT id INTO massage_id FROM category WHERE name = 'Массаж и SPA';
    SELECT id INTO epil_id    FROM category WHERE name = 'Эпиляция и депиляция';
    SELECT id INTO tattoo_id  FROM category WHERE name = 'Тату и перманент';

    cat_ids   := ARRAY[nail_id, hair_id, barber_id, brows_id, makeup_id,
                       cosmo_id, massage_id, epil_id, tattoo_id];
    cat_slugs := ARRAY['nail','hair','barber','brows','makeup',
                       'cosmo','massage','epil','tattoo'];

    FOR cat_idx IN 1..9 LOOP
        cat_id   := cat_ids[cat_idx];
        cat_slug := cat_slugs[cat_idx];

        FOR i IN 1..25 LOOP
            new_uid   := uuid_generate_v4();
            city_i    := ((cat_idx * 7 + i * 3) % 15) + 1;
            street_i  := ((cat_idx * 5 + i * 2) % 15) + 1;
            exp_years := 2 + ((cat_idx + i) % 14);

            -- barber: преимущественно мужчины; tattoo: смешанно; остальные: преимущественно женщины
            is_female := CASE cat_idx
                WHEN 3 THEN (i % 8 = 0)
                WHEN 9 THEN (i % 3 = 0)
                ELSE        (i % 5 != 0)
            END;

            fname := CASE WHEN is_female THEN female_firsts[((i - 1) % 25) + 1]
                                         ELSE male_firsts  [((i - 1) % 25) + 1] END;
            lname := CASE WHEN is_female THEN female_lasts [((i - 1) % 25) + 1]
                                         ELSE male_lasts   [((i - 1) % 25) + 1] END;

            rating_v := GREATEST(3.5, LEAST(5.0, ROUND((4.80 - (i - 1) * 0.05)::NUMERIC, 2)));
            review_v := GREATEST(5,   500 - (i - 1) * 18);

            bio_text := CASE cat_idx
                WHEN 1 THEN
                    'Мастер маникюра и педикюра с ' || exp_years || '-летним опытом. ' ||
                    'Работаю с гель-лаком, акрилом и биогелем. Создаю авторский дизайн любой сложности.'
                WHEN 2 THEN
                    'Парикмахер-стилист с ' || exp_years || '-летним опытом. ' ||
                    'Специализируюсь на стрижках, окрашивании и уходовых процедурах для волос любого типа.'
                WHEN 3 THEN
                    'Профессиональный барбер с ' || exp_years || '-летним стажем. ' ||
                    'Классические и современные мужские стрижки, оформление бороды и усов.'
                WHEN 4 THEN
                    'Мастер по бровям и ресницам с ' || exp_years || '-летним опытом. ' ||
                    'Архитектура бровей, ламинирование, микроблейдинг и наращивание ресниц.'
                WHEN 5 THEN
                    'Профессиональный визажист с ' || exp_years || '-летним опытом. ' ||
                    'Свадебный, вечерний и дневной макияж. Выезд на дом и площадку.'
                WHEN 6 THEN
                    'Косметолог-эстетист с ' || exp_years || '-летним опытом. ' ||
                    'Уходовые процедуры, пилинги, мезотерапия и аппаратная косметология.'
                WHEN 7 THEN
                    'Массажист с ' || exp_years || '-летним опытом. ' ||
                    'Расслабляющий, лечебный и антицеллюлитный массаж, SPA-программы.'
                WHEN 8 THEN
                    'Специалист по эпиляции с ' || exp_years || '-летним опытом. ' ||
                    'Восковая эпиляция, шугаринг и биоэпиляция любых зон.'
                WHEN 9 THEN
                    'Тату-мастер с ' || exp_years || '-летним опытом. ' ||
                    CASE i % 3
                        WHEN 0 THEN 'Специализируюсь на реализме и акварели.'
                        WHEN 1 THEN 'Геометрия, минимализм и дотворк.'
                        ELSE        'Перманентный макияж бровей, губ и стрелок.'
                    END
            END;

            INSERT INTO "user" (user_id, email, password_hash, role)
            VALUES (
                new_uid,
                'master_' || cat_slug || '_' || i || '@seed.okoshki.ru',
                pwd_hash,
                'master'
            );

            INSERT INTO masters (
                user_id, category_id, first_name, last_name,
                address, city, bio, timezone, lat, lon,
                rating, review_count
            ) VALUES (
                new_uid,
                cat_id,
                fname,
                lname,
                streets[street_i] || ', д. ' || ((cat_idx * 11 + i * 7) % 120 + 1),
                cities[city_i],
                bio_text,
                'Europe/Moscow',
                city_lats[city_i] + (((cat_idx * 3 + i * 2) % 21) - 10) * 0.005,
                city_lons[city_i] + (((cat_idx * 4 + i * 3) % 21) - 10) * 0.005,
                rating_v,
                review_v
            );
        END LOOP;
    END LOOP;
END $$;
