# Текущее состояние

> Описывает фактически реализованное на момент написания. Решения о том,
> как закрывать перечисленные ниже ограничения, зафиксированы в
> `docs/analysis.md` → «Принятые решения». Этот документ обновится после
> реализации.



## Что уже реализовано

### Резолвер XML-шаблонов (`internal/tmpl/`)

- **Парсер XML → DOM-подобное дерево `Element{Name, Attrs, Text, Children}`.**
  Файл: `parser.go`. Использует `encoding/xml` с `Strict=false`.
- **Глобальный индекс `basename → absPath`.** Файл: `index.go`. Однократно
  обходит всё дерево `templates/` через `filepath.WalkDir`. Для родительских
  ссылок со слешем (`parent="civ/athen"`) делает suffix-lookup среди всех
  путей.
- **Резолвер с кешем.** Файл: `resolver.go`. Для одного шаблона:
  - Читает `parent="a|b|c"`, разбивает по `|`, рекурсивно резолвит каждого
    родителя, потом мержит между собой слева направо, потом мержит с самим
    шаблоном.
  - Защита от циклов (`stack` map).
  - Кеш итоговых разрешённых сущностей (`map[absPath]*Element`).
- **Merge с особой семантикой 0 A.D.** Файл: `merge.go`.
  - `replace=""` → ребёнок полностью заменяет родительские дочерние элементы.
  - `datatype="tokens"` → токены конкатенируются; токен с префиксом `-`
    удаляется из унаследованного списка.
  - `op="mul"` / `op="add"` на числовом листе → арифметика с родительским
    значением.
  - Иначе — рекурсивное слияние одноимённых дочерних элементов; текст
    ребёнка перекрывает родительский.
- **Подстановки `{civ}` / `{native}`.** Файл: `tokens.go`. Применяются
  во время рендера, не во время резолва.

**Тесты:** `resolver_test.go`. Проверяют:
- `units/spart/infantry_spearman_b.xml` → HP=100, Hack=4.5, Pierce=4, wood=50;
- миксин `hoplite` добавляет `phalanx` к `Formations`;
- `op="mul"` на cavalry даёт WalkSpeed = 9 × 2 = 18;
- `-structures/{civ}/wallset_stone` удаляет токен из `Builder/Entities`.

### Загрузчик цивилизации (`internal/civdata/`)

- **15 цив с алиасами.** Файл: `codes.go`. Структура `CivCode{Code, NameEN,
  NameRU, OutputFile, Aliases}`. `ResolveCivInput()` — case-insensitive
  substring match сначала по коду, потом по алиасам.
- **Парсинг `civs/{code}.json`.** Файл: `civ.go`. Поля `Code`, `Culture`
  (через `RawMessage` — может быть `string` или `[string]`), `CivBonuses`,
  `TeamBonuses`, `WallSets`, `StartEntities`, `SkirmishReplacements`,
  `SelectableInGameSetup`, `AINames`. Поля `Emblem`, `Name`, `History`,
  `Region`, `Music` **не парсятся**.
- **Перечисление зданий и юнитов.** Файл: `enumerate.go`.
  - `Buildings(dir, civ, resolver)` — `filepath.Glob("structures/{civ}/*.xml")`.
  - `Units(dir, civ, resolver)` — то же для юнитов.
  - Классификаторы по basename: `IsHero`, `IsChampion`, `IsSupport`,
    `IsShip`, `IsSiege`, `IsCatafalque`, `IsHealer`.
  - Группировка по фазам через `Identity/RequiredTechnology` (захардкожен
    список фаза-кодов на 17 цив для town/city, плюс fallback по префиксу).
  - Фиксированный порядок зданий внутри фазы (`buildingOrderHints` — 26
    элементов от `civil_centre` до `wonder`).

**Тесты:** `codes_test.go`. Проверяют резолюцию `spart`, `спартанцы`,
`German`, `Germans`, `Афиняне`, `Han`, `карфагеняне`, неизвестная строка,
пустая строка.

### Технологии (`internal/tech/tech.go`)

- **Структура `Technology`.** Поля: `GenericName`, `Description`,
  `SpecificName` (`map[string]any`), `AutoResearch`, `Cost`, `ResearchTime`,
  `Tooltip`, `Modifications` (с `Value`/`Multiply`/`Add`/`Replace`),
  `Affects`, `Requirements`, `Supersedes`, `Pair`, `Top`, `Bottom`, `Icon`,
  `ReplacedBy`. Поля `Replaces` (массив), `RequirementsTooltip` —
  отсутствуют.
- **`Catalog.ByName(name)`** — лениво подгружает один JSON по имени, ищет
  в корне `technologies/` и в `civbonuses/`. Кеш в памяти.
- **`Catalog.AllCivBonuses(civ)`** — сканирует `civbonuses/`, фильтрует
  по `RequiresCiv(req)`. Возвращает только техи с `requirements.civ == civ`.
- **`Catalog.AllNotCiv(civ)`** — сканирует все `technologies/*.json`
  (рекурсивно), возвращает те, где `NotCivList(req)` содержит код цивы.
- **`NotCivList(req)`** — рекурсивно собирает все `notciv` (включая
  обёрнутые в `all`/`any`).
- **`RequiresCiv(req)`** — извлекает `civ` из верхнего уровня или из
  первого вложения в `all`.

### Авры (`internal/aura/aura.go`)

- **Структура `Aura`.** Поля: `Type`, `Radius`, `AffectedPlayers`, `Affects`
  (`[]any` — может быть строка или массив строк), `Modifications`,
  `AuraName`, `AuraDescription`.
- **`ListInDir(dir, prefix)`** — глоб JSON в директории с фильтром по
  префиксу basename. Используется для `auras/units/heroes/{civ}_hero_*` и
  `auras/units/catafalques/{civ}_*`.
- **Авры зданий** (`auras/structures/`) — структура папки видна, но
  загрузчик их не сканирует.
- **Авры игрока** (`auras/players/`, `auras/teambonuses/`) — не
  обрабатываются.

### i18n (`internal/i18n/`)

- **`ru.go`** — `ResourceName(food/wood/stone/metal)`, `DamageType(Hack/
  Pierce/Crush/Capture/Fire)`, `PhaseRequirement(phase_*)`, `FormatNumber`,
  `FormatPercent`.
- **`modifier.go`** — `DescribeModification(m)`/`DescribeModifications(mods)`
  + `pathTranslations` (~50 ключей) для перевода путей вида `Health/Max`,
  `Attack/Melee/Damage/Hack`, `ResourceGatherer/Rates/food.meat` и т.п.
  Незнакомый путь возвращается как есть.

### Рендер (`internal/render/`)

- **`Generator`** в `report.go` — единая точка входа `Generate(civInfo)`.
  Загружает civ JSON, здания, юниты, бонусы, blacklist, ауры героев и
  катафалка; вызывает методы рендера по очереди.
- **`renderHeader`** — заголовок и инфо-блок.
- **`renderOverview`** — код, культура, стартовые юниты, командный бонус,
  таблица цивилизационных бонусов (объединение `civ.CivBonuses` и
  `AllCivBonuses`), список `notciv`-blacklist.
- **`renderPhases`** — три блока (`VILLAGE`/`TOWN`/`CITY`), каждый с
  отсортированными зданиями.
- **`renderBuilding`** — таблица параметров здания + подтаблицы «Тренирует»
  (фильтрует юниты, которых нет в `units/{civ}/`) и «Исследует» (с подгрузкой
  каждого тех-JSON через `Catalog.ByName`).
- **`renderUnitsDetail`** — приложение, классификация юнитов на 10 групп.
  Для героев — таблица аур по basename hero-файла и токенам в `Auras`.
  Для катафалка — таблица аур из `auras/units/catafalques/`.
- **`renderSummary`** — сводная таблица в конце.
- **`format.go`** — форматтеры одного поля шаблона: `FormatCost`,
  `FormatBuildTime`, `FormatHP`, `FormatArmor`, `FormatArmorHPC`,
  `FormatVision`, `FormatGarrison`, `FormatPopulationBonus`,
  `FormatTerritory`, `FormatMeleeAttack`, `FormatRangedAttack`,
  `FormatAttackShort`, `FormatWalkSpeed`, `FormatPopulation`,
  `FormatGenericName`.

**Тесты:** `numeric_test.go` (точечная проверка значений после резолва),
`golden_test.go` (генерация отчёта по `germ`, проверка длины ≥ 700 строк
и наличия обязательных секций).

### CLI (`cmd/civreport/main.go`)

- **Флаги:** `--gamedata`, `--out`, `--all`, `--check`.
- **Резолюция пути:** `--gamedata` → `OAD_GAMEDATA_ROOT` → дефолт.
- **Режимы:**
  - одиночный: `civreport <alias>` пишет файл по `OutputFile` из
    таблицы `civdata.Civilizations`;
  - `--all`: рендер всех 15 цив подряд (продолжает при ошибках);
  - `--check`: smoke-тест, печатает строки и не пишет файлов.

## Архитектура

```
       вход                                          выход
       ─────                                         ─────
   civreport spart                              spartans_buildings_report.md
        │                                                ▲
        ▼                                                │
   ┌──────────────────┐                                  │
   │ cmd/civreport    │  alias-резолюция,                │
   │ main.go          │  выбор режима                    │
   └────────┬─────────┘                                  │
            │ civdata.CivCode{Code, NameEN, NameRU, ...} │
            ▼                                            │
   ┌──────────────────┐                                  │
   │ render.Generator │                                  │
   │ Generate()       │                                  │
   └────────┬─────────┘                                  │
            │                                            │
   ┌────────┴──────────────────────────────────────┐     │
   │ Загрузка данных:                              │     │
   │  ├─ civdata.LoadCiv(civs/{code}.json)         │     │
   │  ├─ civdata.Buildings(structures/{civ}/*.xml) │  →  │ → renderHeader
   │  │   └─ tmpl.Resolver.Resolve(path)           │  →  │ → renderOverview
   │  │       └─ parser → index → merge            │  →  │ → renderPhases
   │  ├─ civdata.Units(units/{civ}/*.xml)          │  →  │ → renderUnitsDetail
   │  ├─ tech.Catalog.AllCivBonuses(civ)           │  →  │ → renderSummary
   │  ├─ tech.Catalog.AllNotCiv(civ)               │     │
   │  ├─ aura.ListInDir(auras/.../heroes/...)      │     │
   │  └─ aura.ListInDir(auras/.../catafalques/...) │     │
   └───────────────────────────────────────────────┘     │
                                                         │
              ┌────────────────────────────┐             │
              │ strings.Builder            │ ────────────┘
              │ (markdown пишется через    │
              │  fmt.Fprintln/Fprintf)     │
              └────────────────────────────┘
```

Поток: вход → резолюция кода цивы → один проход через `Generate` →
последовательный вызов рендер-методов, каждый дёргает свои данные через
ленивые загрузчики (`Catalog.ByName` подгружает индивидуальные техи
по требованию из `renderResearches`) → сборка строки в `strings.Builder` →
запись на диск.

## Известные ограничения

### Покрыто, но обрезано

- **Таблица «Исследует» под зданием.** Если тех ссылается на парный
  (`pair_*`-файл), показывается строка с пустыми полями (нет cost/tooltip
  в самом pair-файле, top/bottom не разворачиваются).
- **Колонка «Фаза» в таблице техов.** Парсится только `req.tech` и первый
  `tech` в `req.all`. Сложные конструкции (`req.entity`, `req.any`) → «—».
- **Авры героев.** Матчатся либо по basename hero-файла (`hero_leonidas` →
  `spart_hero_leonidas_*`), либо по токенам в `<Auras>` шаблона. Авры
  зданий (Wonder, Temple) — не выводятся.
- **Подтаблицы «Тренирует».** Фильтрация: если токен после `{civ}`-
  подстановки не находится в `units/{civ}/`, строка пропускается. Это
  правильно для «мёртвых» шаблонов, но также скрывает юнитов, которых
  можно было бы построить через цепочку (если бы их XML лежали где-то
  вне `units/{civ}/`).

### Не покрыто

- **Парные технологии** (`pair_*.json`) — формально не разворачиваются;
  поля `Top`/`Bottom`/`Pair` в `tech.Technology` есть, но рендер их не
  использует.
- **Глобальные `autoResearch`-техи** (`unit_advanced.json`, `unit_elite.json`)
  — не сканируются, в отчёт не попадают.
- **Player-bound auras** (`auras/players/`, `auras/teambonuses/`) — не
  загружаются.
- **Авры зданий** (`auras/structures/`) — не загружаются и не
  ассоциируются с зданиями через `<Auras>` token-список.
- **Локализация имён** — `Identity/SpecificName` и `GenericName` берутся
  из шаблона как есть, без перевода. `civ.json.Name`/`History`/`Emblem` —
  не парсятся.
- **`SkirmishReplacements`** — поле парсится, но не применяется (вход в
  отчёт берётся из шаблонов цивы, не из skirmish-вариантов).
- **WallSet** — отдельные `wall_short`/`wall_medium`/`wall_long`/`wall_gate`/
  `wall_tower` рендерятся как самостоятельные здания, без группировки в
  один блок «стены».
- **Транзитивное замыкание от StartEntities** — игра в `gui/reference/`
  идёт от стартовых сущностей через `Trainer`/`Builder`/`Trainer/Technologies`.
  У нас — `filepath.Glob` по папке цивы. Различия для большинства цив
  минимальны, но не нулевые.
- **Бонусы атаки** (`Bonuses/Bonus*` в `<Attack>`) — `+2.5× vs Cavalry`
  у копейщиков и т.п. не рендерятся.
- **Splash damage**, **capture details**, **status effects** в атаке —
  не рендерятся.
- **Loot**, **Upkeep**, **Treasure**, **Resource trickle**, **Resource supply**
  (запас на полях/деревьях), **Resource dropsite labels** — не рендерятся.
- **Healer** (range/HP/repeat) — не выделяется отдельной таблицей,
  лекари идут как обычные юниты с базовыми параметрами.
- **Turrets** и **projectiles из зданий** (`+N стрел за гарнизонного
  юнита`) — не рендерятся.
- **Run speed** — выводится только `WalkSpeed`, `RunMultiplier`/`RunSpeed`
  — нет.
- **Capture resistance**, **Status effect resistance** — не рендерятся.
- **`specificName`** технологий и фаз — не используется при рендере.
- **`supersedes`/`replaces`/`replacedBy`** — храним частично (Replaces
  вообще нет), при рендере не используем.
- **`requirementsTooltip`** на технологиях — не парсится и не показывается.

### Не покрыто тестами

- Парсинг `civ.json` (нет unit-тестов, есть только эффект через golden).
- Загрузка технологий и blacklist (`AllNotCiv` / `AllCivBonuses`) — нет
  unit-тестов на конкретные числа/списки.
- Авры — нет unit-тестов.
- i18n — нет unit-тестов на `pathTranslations`.
- Рендер форматтеров (`format.go`) — нет unit-тестов; покрыто косвенно
  через `golden_test.go`.
- Golden-тест для `germ` — структурный (наличие секций, длина ≥ 700),
  не сравнивает с эталоном по содержимому.
- Smoke-проверка всех 15 цив (`--check`) — есть в CLI, но не оформлена
  как `go test`.
