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
- **Транзитивное замыкание от StartEntities.** Файл: `reach.go`.
  `Reach(civ, idx, resolver, catalog)` — единственный источник истины
  достижимых зданий/юнитов/техов. Идёт от `StartEntities` цивы по рёбрам
  `Trainer/Entities`, `Builder/Entities`, `Researcher/Technologies` до
  fixed point. `ReachResult` содержит `Buildings`, `Units`, `Techs`,
  `WallSets`, `Skipped` (для отладки). `SkipNote{Token, Reason}` —
  диагностическая запись пропущенного шаблона.
  `Buildings()`/`Units()` glob-функции **удалены**.
- **Группировка WallSet.** Файл: `wallset.go`.
  `IdentifyWallSets(buildings, civCode)` — выделяет wallset-обёртки и
  их pieces из общего списка зданий. Возвращает `[]*WallSetGroup`,
  каждый с `Name`, `WrapperEl`, `Pieces []WallPiece`. Wallset-обёртки
  и составные части исключаются из `ReachResult.Buildings`.
- **Перечисление/классификация юнитов.** Файл: `enumerate.go`.
  - Классификаторы по basename: `IsHero`, `IsChampion`, `IsSupport`,
    `IsShip`, `IsSiege`, `IsCatafalque`, `IsHealer`.
  - Группировка по фазам через `Identity/RequiredTechnology` (захардкожен
    список фаза-кодов на 17 цив для town/city, плюс fallback по префиксу).
  - Фиксированный порядок зданий внутри фазы (`buildingOrderHints` — 26
    элементов от `civil_centre` до `wonder`).

**Тесты:** `codes_test.go`. Проверяют резолюцию `spart`, `спартанцы`,
`German`, `Germans`, `Афиняне`, `Han`, `карфагеняне`, неизвестная строка,
пустая строка. `reach_test.go`, `wallset_test.go` — интеграционные тесты
с реальными данными 0ad (пропускаются если данные недоступны).

### Технологии (`internal/tech/`)

- **Структура `Technology`.** Поля: `GenericName`, `Description`,
  `SpecificName` (`map[string]any`), `AutoResearch`, `Cost`, `ResearchTime`,
  `Tooltip`, `Modifications` (с `Value`/`Multiply`/`Add`/`Replace`),
  `Affects`, `Requirements`, `Supersedes`, `Pair`, `Top`, `Bottom`, `Icon`,
  `ReplacedBy`. Поля `Replaces` (массив), `RequirementsTooltip` —
  отсутствуют.
- **`Catalog.ByName(name)`** — лениво подгружает один JSON по имени, ищет
  в корне `technologies/` и в `civbonuses/`. Кеш в памяти.
- **`Catalog.LoadAll()`** — загружает все технологии разом (рекурсивный
  обход `technologies/`). **`Catalog.AllLoaded()`** — срез всех
  загруженных техов (нужен для построения `Index`).
- **`Catalog.AllCivBonuses(civ)`** — сканирует `civbonuses/`, фильтрует
  по `RequiresCiv(req)`. Возвращает только техи с `requirements.civ == civ`.
- **`Catalog.AllNotCiv(civ)`** — сканирует все `technologies/*.json`
  (рекурсивно), возвращает те, где `NotCivList(req)` содержит код цивы.
- **`NotCivList(req)`** — рекурсивно собирает все `notciv` (включая
  обёрнутые в `all`/`any`).
- **`RequiresCiv(req)`** — извлекает `civ` из верхнего уровня или из
  первого вложения в `all`.
- **`ExpandPair(catalog, name)`** (файл: `pair.go`) — разворачивает
  pair-обёртку в два отдельных теха (`top`, `bottom`). Если `name` не
  является pair-техом, возвращает `ok=false`.
- **`NewIndex(catalog)`** (файл: `replaces.go`) — строит граф
  replaces/supersedes по всем техам (`Catalog.LoadAll` + `AllLoaded`).
  Методы: `Index.ResolveForCiv(name, civ)` — возвращает civ-специфичный
  вариант теха (или generic); `Index.Chain(name)` — возвращает `ChainInfo`
  со списками `Supersedes`, `Replaces`, `SupersededBy`, `ReplacedBy` для
  построения chain-суффикса. `Index.Warnings` — диагностика дублей.

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
  Pierce/Crush/Capture/Fire/Poison)`, `PhaseRequirement(phase_*)`, `FormatNumber`,
  `FormatPercent`.
- **`modifier.go`** — `DescribeModification(m)`/`DescribeModifications(mods)`
  + `pathTranslations` (~50 ключей) для перевода путей вида `Health/Max`,
  `Attack/Melee/Damage/Hack`, `ResourceGatherer/Rates/food.meat` и т.п.
  Незнакомый путь возвращается как есть.

### Рендер (`internal/render/`)

- **`Generator`** в `report.go` — единая точка входа `Generate(civInfo)`.
  Использует `civdata.Reach(...)` как единственный источник списков
  зданий/юнитов/техов. Поле `Generator.Index *tech.Index` лениво
  строится при первом вызове `Generate` через `tech.NewIndex(catalog)`.
- **`renderHeader`** — заголовок и инфо-блок.
- **`renderOverview`** — код, культура, стартовые юниты, командный бонус,
  таблица цивилизационных бонусов (объединение `civ.CivBonuses` и
  `AllCivBonuses`), список `notciv`-blacklist.
- **`renderPhases`** — три блока (`VILLAGE`/`TOWN`/`CITY`), каждый с
  отсортированными зданиями. Wallset выводится отдельным блоком `### Стены`
  с 7-колоночной таблицей pieces через `renderWallSetBlock` (`wallset.go`).
- **`renderBuilding`** — таблица параметров здания + подтаблицы «Тренирует»
  и «Исследует». В «Исследует»: pair-технологии разворачиваются в две строки
  с маркером `◐ — парная (выбрать одно)` (`pair.go`); каждая строка теха
  содержит chain-суффикс «(заменяет: X; апгрейд от Y; ...)» через
  `chainSuffix`. Фаза в колонке «Фаза» резолвится через
  `requirementPhase(t, civ, idx)` с поддержкой `Index.ResolveForCiv`
  для civ-вариантов фаз (Афины → `phase_town_athen`, Персы → `phase_town_pers`).
- **`renderUnitsDetail`** — приложение, классификация юнитов на 10 групп.
  Для героев — таблица аур по basename hero-файла и токенам в `Auras`.
  Для катафалка — таблица аур из `auras/units/catafalques/`.
- **`renderUnitBlock`** (после эпика 4a) — детальная 2-колоночная
  таблица параметров юнита включает: `Сопротивление захвату`,
  `Сопротивление статус-эффектам` (если поля присутствуют) после
  строки `Броня (H/P/C)`; для каждого attack-mode (`Melee`/`Ranged`/
  `Capture`) — строки `Атака (m)`, `Бонусы (m)`, `Предпочитает (m)`,
  `Брызги (m)`, `Накладывает (m) × N` (по строке на каждый
  ApplyStatus). `Атака (захват)` пишется всегда, когда у юнита есть
  `<Attack><Capture>` (даже унаследованный дефолт). Тренировочная
  таблица (`renderTrains`) остаётся компактной 9-колоночной.
- **`renderSummary`** — сводная таблица в конце.
- **`format.go`** — форматтеры одного поля шаблона: `FormatCost`,
  `FormatBuildTime`, `FormatHP`, `FormatArmor`, `FormatArmorHPC`,
  `FormatVision`, `FormatGarrison`, `FormatPopulationBonus`,
  `FormatTerritory`, `FormatMeleeAttack`, `FormatRangedAttack`,
  `FormatAttackShort`, `FormatWalkSpeed`, `FormatPopulation`,
  `FormatGenericName`, `FormatCaptureResistance`,
  `FormatStatusEffectResistance` (эпик 4a).
  Внутренние помощники для attack-mode: `formatAttackBonuses`,
  `formatPreferredClasses`, `formatSplash`, `formatCaptureAttack`,
  `formatApplyStatuses(modeEl, byCode)`.

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
   civreport spart                    spartans_overview.md + spartans_structree.md
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
   │  ├─ civdata.Reach(civ, idx, resolver, catalog)│  →  │ → renderHeader
   │  │   └─ transitive closure from StartEntities │  →  │ → renderOverview
   │  │       └─ Buildings/Units/WallSets/Techs    │  →  │ → renderPhases
   │  │                                            │  →  │ → renderUnitsDetail
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

- **Колонка «Фаза» в таблице техов.** `requirementPhase` теперь
  поддерживает `req.entity` (через `Index.ResolveForCiv`) и первый
  `tech` в `req.all`. Конструкция `req.any` и смешанные ветки
  `all(tech+notciv)` → «—».
- **Авры героев.** Матчатся либо по basename hero-файла (`hero_leonidas` →
  `spart_hero_leonidas_*`), либо по токенам в `<Auras>` шаблона. Авры
  зданий (Wonder, Temple) — не выводятся.

### Не покрыто

- **Глобальные `autoResearch`-техи** (`unit_advanced.json`, `unit_elite.json`)
  — не сканируются в отчёт как самостоятельные строки (описание в
  `common.md` есть, но к таблицам юнитов не привязано).
- **Player-bound auras** (`auras/players/`, `auras/teambonuses/`) — не
  загружаются.
- **Авры зданий** (`auras/structures/`) — не загружаются и не
  ассоциируются с зданиями через `<Auras>` token-список.
- **Локализация имён** — `Identity/SpecificName` и `GenericName` берутся
  из шаблона как есть, без перевода. `civ.json.Name`/`History`/`Emblem` —
  не парсятся.
- **`SkirmishReplacements`** — поле парсится, но не применяется (вход в
  отчёт берётся из шаблонов цивы, не из skirmish-вариантов).
- **Бонусы атаки** ✓ закрыто эпиком 4a — `×2.5 vs Cavalry`
  у копейщиков рендерится строкой `Бонусы (ближ.)`.
- **Splash damage**, **capture details**, **status effects** в атаке
  ✓ закрыто эпиком 4a — `Брызги`, `Атака (захват)`, `Накладывает`
  per attack-mode.
- **Loot**, **Upkeep**, **Treasure**, **Resource trickle**, **Resource supply**
  (запас на полях/деревьях), **Resource dropsite labels** — не рендерятся.
- **Healer** (range/HP/repeat) — не выделяется отдельной таблицей,
  лекари идут как обычные юниты с базовыми параметрами.
- **Turrets** и **projectiles из зданий** (`+N стрел за гарнизонного
  юнита`) — не рендерятся.
- **Run speed** — выводится только `WalkSpeed`, `RunMultiplier`/`RunSpeed`
  — нет.
- **Capture resistance**, **Status effect resistance** ✓ закрыто
  эпиком 4a — `FormatCaptureResistance`,
  `FormatStatusEffectResistance` (в R28 эти поля отсутствуют в
  шаблонах, но логика готова к появлению).
- **`specificName`** технологий и фаз — не используется при рендере.
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
