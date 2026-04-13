DO $$
DECLARE
    -- Переменные для хранения ID корневых категорий (Уровень 1)
    nail_id UUID := uuid_generate_v4();
    hair_id UUID := uuid_generate_v4();
    barber_id UUID := uuid_generate_v4();
    brows_id UUID := uuid_generate_v4();
    makeup_id UUID := uuid_generate_v4();
    cosmo_id UUID := uuid_generate_v4();
    massage_id UUID := uuid_generate_v4();
    epil_id UUID := uuid_generate_v4();
    tattoo_id UUID := uuid_generate_v4();

    -- Переменные для хранения ID подкатегорий (Уровень 2), у которых будут свои дети
    manicure_id UUID := uuid_generate_v4();
    pedicure_id UUID := uuid_generate_v4();
    podology_id UUID := uuid_generate_v4();
    color_id UUID := uuid_generate_v4();
    care_id UUID := uuid_generate_v4();
    style_id UUID := uuid_generate_v4();
    brow_shape_id UUID := uuid_generate_v4();
    lash_ext_id UUID := uuid_generate_v4();
    cosmo_est_id UUID := uuid_generate_v4();
    cosmo_app_id UUID := uuid_generate_v4();
    cosmo_inj_id UUID := uuid_generate_v4();
    cosmo_mas_id UUID := uuid_generate_v4();
    perm_id UUID := uuid_generate_v4();
BEGIN
    -- УРОВЕНЬ 1 (Корневые папки)
    INSERT INTO category (id, name, description) VALUES
        (nail_id, 'Ногтевой сервис', 'Всё для красоты ваших рук и ног'),
        (hair_id, 'Волосы', 'Стрижки, укладки, уход и окрашивание'),
        (barber_id, 'Барбершоп', 'Мужские стрижки и оформление бороды'),
        (brows_id, 'Брови и ресницы', 'Архитектура, ламинирование, наращивание'),
        (makeup_id, 'Макияж', 'Дневной, вечерний и свадебный макияж'),
        (cosmo_id, 'Косметология', 'Эстетика, инъекции и аппаратные процедуры'),
        (massage_id, 'Массаж и SPA', 'Расслабление и оздоровление тела'),
        (epil_id, 'Эпиляция и депиляция', 'Гладкая кожа надолго'),
        (tattoo_id, 'Тату и перманент', 'Перманентный макияж и татуировки');

    -- УРОВЕНЬ 2 и 3: Ногтевой сервис
    INSERT INTO category (id, parent_id, name) VALUES
        (manicure_id, nail_id, 'Маникюр'),
        (uuid_generate_v4(), manicure_id, 'Классический маникюр'),
        (uuid_generate_v4(), manicure_id, 'Аппаратный маникюр'),
        (uuid_generate_v4(), manicure_id, 'Комбинированный маникюр'),

        (pedicure_id, nail_id, 'Педикюр'),
        (uuid_generate_v4(), pedicure_id, 'Smart-педикюр'),
        (uuid_generate_v4(), pedicure_id, 'Кислотный педикюр'),
        (uuid_generate_v4(), pedicure_id, 'Классический педикюр'),

        (uuid_generate_v4(), nail_id, 'Наращивание ногтей'),

        (podology_id, nail_id, 'Подология'),
        (uuid_generate_v4(), podology_id, 'Медицинский педикюр'),
        (uuid_generate_v4(), podology_id, 'Лечение вросшего ногтя');

    -- УРОВЕНЬ 2 и 3: Волосы
    INSERT INTO category (id, parent_id, name) VALUES
        (uuid_generate_v4(), hair_id, 'Женские стрижки'),

        (color_id, hair_id, 'Окрашивание волос'),
        (uuid_generate_v4(), color_id, 'Окрашивание в тон'),
        (uuid_generate_v4(), color_id, 'Сложное окрашивание (Airtouch, Шатуш, Балаяж)'),

        (care_id, hair_id, 'Уход за волосами'),
        (uuid_generate_v4(), care_id, 'Кератиновое выпрямление'),
        (uuid_generate_v4(), care_id, 'Ботокс для волос'),
        (uuid_generate_v4(), care_id, 'Нанопластика'),

        (style_id, hair_id, 'Укладки и прически'),
        (uuid_generate_v4(), style_id, 'Вечерняя прическа'),
        (uuid_generate_v4(), style_id, 'Свадебная прическа'),

        (uuid_generate_v4(), hair_id, 'Наращивание волос');

    -- УРОВЕНЬ 2: Барбершоп
    INSERT INTO category (id, parent_id, name) VALUES
        (uuid_generate_v4(), barber_id, 'Мужские стрижки'),
        (uuid_generate_v4(), barber_id, 'Стрижка и моделирование бороды'),
        (uuid_generate_v4(), barber_id, 'Камуфляж седины');

    -- УРОВЕНЬ 2 и 3: Брови и ресницы
    INSERT INTO category (id, parent_id, name) VALUES
        (brow_shape_id, brows_id, 'Оформление бровей'),
        (uuid_generate_v4(), brow_shape_id, 'Архитектура бровей'),
        (uuid_generate_v4(), brow_shape_id, 'Окрашивание хной/краской'),

        (uuid_generate_v4(), brows_id, 'Долговременная укладка / Ламинирование бровей'),

        (lash_ext_id, brows_id, 'Наращивание ресниц'),
        (uuid_generate_v4(), lash_ext_id, 'Классическое наращивание'),
        (uuid_generate_v4(), lash_ext_id, 'Объемное наращивание (2D-5D)'),

        (uuid_generate_v4(), brows_id, 'Ламинирование ресниц');

    -- УРОВЕНЬ 2: Макияж
    INSERT INTO category (id, parent_id, name) VALUES
        (uuid_generate_v4(), makeup_id, 'Дневной макияж'),
        (uuid_generate_v4(), makeup_id, 'Вечерний макияж'),
        (uuid_generate_v4(), makeup_id, 'Свадебный макияж'),
        (uuid_generate_v4(), makeup_id, 'Курс «Сам себе визажист»');

    -- УРОВЕНЬ 2 и 3: Косметология
    INSERT INTO category (id, parent_id, name) VALUES
        (cosmo_est_id, cosmo_id, 'Эстетическая косметология'),
        (uuid_generate_v4(), cosmo_est_id, 'Чистка лица'),
        (uuid_generate_v4(), cosmo_est_id, 'Пилинги'),
        (uuid_generate_v4(), cosmo_est_id, 'Маски'),

        (cosmo_app_id, cosmo_id, 'Аппаратная косметология'),
        (uuid_generate_v4(), cosmo_app_id, 'SMAS-лифтинг'),
        (uuid_generate_v4(), cosmo_app_id, 'RF-лифтинг'),
        (uuid_generate_v4(), cosmo_app_id, 'Микротоки'),

        (cosmo_inj_id, cosmo_id, 'Инъекционная косметология'),
        (uuid_generate_v4(), cosmo_inj_id, 'Увеличение губ'),
        (uuid_generate_v4(), cosmo_inj_id, 'Ботулинотерапия (Ботокс)'),
        (uuid_generate_v4(), cosmo_inj_id, 'Мезотерапия и биоревитализация'),

        (cosmo_mas_id, cosmo_id, 'Массаж лица'),
        (uuid_generate_v4(), cosmo_mas_id, 'Скульптурный массаж'),
        (uuid_generate_v4(), cosmo_mas_id, 'Буккальный массаж');

    -- УРОВЕНЬ 2: Массаж и SPA
    INSERT INTO category (id, parent_id, name) VALUES
        (uuid_generate_v4(), massage_id, 'Классический / Оздоровительный массаж'),
        (uuid_generate_v4(), massage_id, 'Антицеллюлитный / Лимфодренажный массаж'),
        (uuid_generate_v4(), massage_id, 'SPA-программы и обертывания'),
        (uuid_generate_v4(), massage_id, 'Массаж спины и ШВЗ');

    -- УРОВЕНЬ 2: Эпиляция и депиляция
    INSERT INTO category (id, parent_id, name) VALUES
        (uuid_generate_v4(), epil_id, 'Шугаринг (сахарная депиляция)'),
        (uuid_generate_v4(), epil_id, 'Восковая депиляция (ваксинг)'),
        (uuid_generate_v4(), epil_id, 'Лазерная эпиляция'),
        (uuid_generate_v4(), epil_id, 'Электроэпиляция');

    -- УРОВЕНЬ 2 и 3: Тату и перманент
    INSERT INTO category (id, parent_id, name) VALUES
        (perm_id, tattoo_id, 'Перманентный макияж'),
        (uuid_generate_v4(), perm_id, 'Перманент бровей'),
        (uuid_generate_v4(), perm_id, 'Перманент губ'),
        (uuid_generate_v4(), perm_id, 'Межресничка'),

        (uuid_generate_v4(), tattoo_id, 'Художественная татуировка'),
        (uuid_generate_v4(), tattoo_id, 'Удаление тату и татуажа (лазер/ремувер)'),
        (uuid_generate_v4(), tattoo_id, 'Пирсинг');
END $$;