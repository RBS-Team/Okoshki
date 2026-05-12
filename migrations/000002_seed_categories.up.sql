DO $$
DECLARE
    nail_id    UUID := uuid_generate_v4();
    hair_id    UUID := uuid_generate_v4();
    barber_id  UUID := uuid_generate_v4();
    brows_id   UUID := uuid_generate_v4();
    makeup_id  UUID := uuid_generate_v4();
    cosmo_id   UUID := uuid_generate_v4();
    massage_id UUID := uuid_generate_v4();
    epil_id    UUID := uuid_generate_v4();
    tattoo_id  UUID := uuid_generate_v4();

BEGIN
    INSERT INTO category (id, name, description) VALUES
        (nail_id,    'Ногтевой сервис',      'Всё для красоты ваших рук и ног'),
        (hair_id,    'Волосы',               'Стрижки, укладки, уход и окрашивание'),
        (barber_id,  'Барбершоп',            'Мужские стрижки и оформление бороды'),
        (brows_id,   'Брови и ресницы',      'Архитектура, ламинирование, наращивание'),
        (makeup_id,  'Макияж',               'Дневной, вечерний и свадебный макияж'),
        (cosmo_id,   'Косметология',         'Эстетика, инъекции и аппаратные процедуры'),
        (massage_id, 'Массаж и SPA',         'Расслабление и оздоровление тела'),
        (epil_id,    'Эпиляция и депиляция', 'Гладкая кожа надолго'),
        (tattoo_id,  'Тату и перманент',     'Перманентный макияж и татуировки');
END $$;
