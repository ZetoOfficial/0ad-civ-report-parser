# CLAUDE.md

## Проект

Детерминированный CLI на Go, генерирующий русскоязычные markdown-отчёты
о цивилизациях 0 A.D. (движок Pyrogenesis) напрямую из XML/JSON шаблонов
игры — замена для LLM-скилла `/civ-report`, который делает то же самое через
парсинг шаблонов с участием модели.

## Стек

- **Go**: 1.25.5 (см. `go.mod`).
- **Зависимости**: только stdlib. `go.mod` пустой по `require`. Внешних
  библиотек на момент написания нет.
- **Сборка**: `Makefile` в корне, цели `build`, `<civ>`, `all-civs`, `check`,
  `test`, `clean`.

## Цель

На входе: код или алиас цивилизации (`spart`, `спарт`, `Spartans`, ...).
На выходе **два** файла на циву (зеркало двух экранов в игре —
Civilization Overview и Structure Tree):

### `<civ>_overview.md` (Civilization Overview)

1. Заголовок, эмблема (опционально), культура, код, стартовые юниты.
2. Исторический текст из `Identity/History` Player-шаблона
   `templates/special/players/<civ>.xml` (опционально, по конфигу).
3. Цивилизационные бонусы — объединение трёх источников:
   `civ.json.CivBonuses`, авторесёрч-техи из
   `data/technologies/civbonuses/` (с `requirements.civ == <code>`),
   player-bound civ-ауры (теоретически — в R28 не используются).
4. Командный бонус — описание авры из
   `data/auras/teambonuses/<civ>_player_teambonus.json` со
   структурированными модификациями.
5. Список технологий, **явно** недоступных циве через
   `requirements.notciv`.
6. Краткие списки уникальных зданий (`CivSpecific Structure`),
   героев и уникальных технологий — как в `civinfo` GUI.

### `<civ>_structree.md` (Structure Tree)

1. Поэтапные блоки `VILLAGE PHASE` / `TOWN PHASE` / `CITY PHASE` с
   зданиями; у каждого здания таблица параметров + подтаблицы
   «Тренирует» (юниты) и «Исследует» (технологии). Парные техи —
   две строки с маркером «либо/либо».
2. Приложение «Детальная информация по типам юнитов» с группировкой
   по ролям (поддержка, пехота, конница, чемпионы, герои с аурами,
   осада, корабли, лекари, катафалк). Для каждого семейства — сноска
   об эффекте `unit_advanced`/`unit_elite` (один раз на семейство, не
   на каждый ранг). Status-эффекты атаки (`<ApplyStatus>`) — внутри
   блока атаки. Иконки (`![alt](path)`) — опционально, по конфигу.
3. Приложение «Сводная таблица строимых зданий».

### `common.md` (один на все цивы, дополнительно)

Глобальные данные, общие для всех 15 цив:
- модификаторы `unit_advanced.json` / `unit_elite.json`;
- описание damage types из `data/damage_types/`;
- описание ресурс-типов из `data/resources/`;
- описание status_effects (`burning`, `poisoned`) из
  `data/status_effects/`;
- дефолты `template_unit.xml`, `template_structure.xml` для полей,
  которые большинство наследников не переопределяют.

Цив-отчёты ссылаются на якоря в `common.md` вместо повторения
базовых значений.

### Семантика числовых значений

Числа в отчёте — **базовые** значения шаблонов с применённым
наследованием (`parent`-цепочки, миксины, `op="mul"`/`op="add"`,
`datatype="tokens"` с `-prefix` удалением, подстановки `{civ}`/`{native}`).
Эффекты технологий и аур к статам **не применяются** — это явно
оговорено в info-блоке каждого отчёта.

## Источник данных

```
/Users/zeto/Projects/study/0ad/binaries/data/mods/public
```

**READ-ONLY.** Утилита только читает оттуда. Конкретные подкаталоги — см.
`docs/sources.md`.

Конфигурируется на запуске:
1. Флаг `--gamedata <path>`,
2. Переменная окружения `OAD_GAMEDATA_ROOT`,
3. Захардкоженный дефолт в `internal/paths/paths.go` (= путь выше).

## Структура пакетов

### Текущее (актуально на момент написания)

```
cmd/civreport/main.go              CLI: парсинг флагов, режимы --all/--check/одиночный
internal/paths/                    конфигурация путей до данных игры
internal/tmpl/                     резолвер XML-шаблонов (parent, op=, tokens, mixins)
  ├── element.go                   DOM-подобное дерево (Element)
  ├── parser.go                    XML → Element
  ├── index.go                     basename → path; suffix-lookup для "civ/athen"
  ├── resolver.go                  разрешение parent="a|b|c", кеш, циклы
  ├── merge.go                     слияние parent+child с op=, replace=, tokens
  ├── tokens.go                    подстановка {civ}/{native}
  └── utils.go                     ParseFloatTrim, ParseInt
internal/civdata/                  цивилизации, перечисление сущностей
  ├── codes.go                     таблица 15 цив + русские/английские алиасы
  ├── civ.go                       парсинг civs/{code}.json (Culture как string|array)
  ├── player.go                    парсинг Player-шаблона (Identity/History/Icon/GenericName)
  ├── enumerate.go                 классификация юнитов, фазы (Buildings/Units glob УДАЛЕНЫ)
  ├── reach.go                     Reach(civ, idx, resolver, catalog) → ReachResult; единственный
  │                                источник достижимых buildings/units/techs от StartEntities
  └── wallset.go                   IdentifyWallSets(buildings, civCode) → []*WallSetGroup; WallPiece
internal/tech/
  ├── tech.go                      технологии: загрузка JSON, AllCivBonuses, AllNotCiv,
  │                                LoadAll(), AllLoaded()
  ├── auto.go                      глобальные autoResearch (unit_advanced/elite)
  ├── pair.go                      ExpandPair(catalog, name) → top, bottom, ok
  ├── replaces.go                  NewIndex(catalog), Index{Chain, ResolveForCiv}, ChainInfo
  └── requirements.go              парсинг entity/all/any/notciv с человеко-описанием
internal/aura/
  ├── aura.go                      авры: hero/catafalque/structure (только список)
  └── teambonus.go                 загрузка auras/teambonuses/<civ>_player_teambonus
internal/i18n/                     русские термины + переводчик путей модификаций
  ├── ru.go                        ResourceName, DamageType, PhaseRequirement, FormatNumber
  ├── modifier.go                  pathTranslations (~50 ключей) + DescribeModification
  └── tech_name.go                 TechDisplayName — человекочитаемые имена технологий
internal/render/                   рендер markdown
  ├── format.go                    форматтеры одного поля (FormatCost, FormatHP, ...)
  ├── report.go                    Generator (+Index *tech.Index), Generate(), header/overview/phases
  ├── overview.go                  рендер блоков для overview-вкладки
  ├── structree.go                 рендер фаз через Reach + WallSet + Pair
  ├── pair.go                      formatTechRow/formatPairRow; chainSuffix; ◐-маркер парных техов
  ├── wallset.go                   renderWallSetBlock; 7-колоночная таблица стен; roleLabel
  ├── common.go                    рендер common.md
  ├── common_data.go               данные для common.md (авто-эффекты, damage, resources)
  ├── units.go                     приложение по юнитам, ауры героев/катафалка
  └── summary.go                   сводная таблица в конце
internal/testutil/
  └── gamedata.go                  helpers для тестов (skipIfNoGamedata, gamedataRoot, newResolver)
testdata/golden/                   эталоны для smoke-тестов
  ├── germans_overview.md          эталон обзорного отчёта по germ (untracked)
  └── germans_structree.md         эталон структурного отчёта по germ (untracked)
```

### Целевая структура (после внедрения решений из `docs/analysis.md`)

```
cmd/civreport/main.go              CLI с расширенными флагами (см. ниже)
config.json                        дефолты конфигурации (рядом с бинарником)
internal/config/config.go          struct Config + загрузка из JSON и CLI
internal/paths/                    без изменений
internal/tmpl/                     без изменений
internal/civdata/
  ├── codes.go                     без изменений
  ├── civ.go                       без изменений (Player-шаблон читается через player.go)
  ├── reach.go                     реализовано в эпике 3
  └── wallset.go                   реализовано в эпике 3
internal/tech/
  ├── tech.go                      + поля RequirementsTooltip, Replaces (массив) — ещё не добавлены
  ├── notciv.go                    NEW: вынести из tech.go
  ├── pair.go                      реализовано в эпике 3
  ├── auto.go                      без изменений
  └── requirements.go              без изменений
internal/aura/
  ├── aura.go                      + структурированный рендер modifications
  └── teambonus.go                 NEW: загрузка auras/teambonuses/<civ>_player_teambonus
internal/i18n/
  ├── ru.go                        без существенных изменений
  ├── modifier.go                  + поддержка per-mod affects
  └── po.go                        NEW: загрузчик l10n/<lang>.public-*.po и lookup
internal/render/
  ├── skeleton_overview.tmpl       text/template для Civilization Overview (NEW — эпик 4+)
  ├── skeleton_structree.tmpl      text/template для Structure Tree (NEW — эпик 4+)
  ├── skeleton_common.tmpl         text/template для common.md (NEW — эпик 4+)
  ├── skeleton.go                  загрузка и применение skeleton-шаблонов (NEW — эпик 4+)
  ├── format.go                    форматтеры новых полей (Loot, Trickle, Turrets,
  │                                Projectiles, Splash, Bonuses, Healer, Run, Capture)
  ├── overview.go                  без изменений (уже реализовано)
  ├── structree.go                 реализовано в эпике 3 (интеграция Reach+Pair+WallSet)
  ├── common.go                    без изменений (уже реализовано)
  └── icons.go                     NEW: опциональная вставка ![alt](path) (эпик 5)
testdata/
  ├── golden/                      эталоны (golden-тест на germ переписать под новый формат)
  └── fixtures/                    минимальный mock simulation/ для unit-тестов
```

Старые `internal/render/{report,units,summary}.go` после рефакторинга
заменяются на `overview.go` / `structree.go` / `common.go`. Их логика
форматирования отдельных полей (`format.go`) переезжает с минимальными
изменениями.

### Примечания к структуре

- **`internal/config/`** — рекомендую отдельный пакет, а не `config.go`
  в корне, чтобы не было циклических импортов с `cmd/`.
- **`internal/i18n/po.go`** загружает `.po` файл по `lang`-коду; lookup
  делается по строке-источнику (msgid) с fallback на оригинал.
- **`skeleton_*.tmpl`** загружаются через `//go:embed` чтобы не
  привязываться к рантайм-путям. Конфиг может позволить переопределить
  путь для разработки.

## Конвенции

- **Тестирование.** Каждый пакет с нетривиальной логикой имеет `*_test.go`
  рядом. Тесты резолвера и numeric spot-checks читают реальные данные 0ad
  и пропускаются (`t.Skipf`), если корень недоступен. Golden-тест на `germ`
  проверяет наличие обязательных секций, но **не** делает byte-diff
  с рукописным эталоном (по решению пользователя — golden только для
  одной цивы).
- **Именование Go.** Стандартные правила. Внутренние пакеты — под `internal/`,
  чтобы они не были импортируемы извне модуля.
- **Что нельзя хардкодить.**
  - Путь к данным игры в коде, кроме дефолта в `paths.go` (можно
    переопределить флагом/env).
  - Список цив — единственное место `internal/civdata/codes.go`. Остальные
    модули берут оттуда.
  - Числовые поля шаблонов — никогда не вводить вручную, всегда читать
    через `tmpl.Resolver` с применённым наследованием.
  - Английские tooltip'ы технологий копируются из JSON как есть; русский
    перевод эффектов делается только через `i18n.pathTranslations` для
    путей модификаций — собственного перевода tooltip-строк нет.
- **Детерминизм.** Все списки сортируются перед рендером (по basename
  для алфавитной сортировки внутри группы, по `BuildingSortKey` для
  фиксированного порядка зданий внутри фазы). Маршрут от ввода до файла
  не использует случайность.
- **После эпика 2**: overview содержит 7 секций + опц. History +
  footer-сноску; common.md наполнен (Advanced/Elite/Auto/Damage/Resources/Status).
- **После эпика 3**: единственный источник истины достижимых сущностей —
  `civdata.Reach(civ, idx, resolver, catalog)`. Старые `Buildings()`/`Units()`
  glob-функции удалены. Pair-техи разворачиваются в две строки (`tech.ExpandPair`).
  Граф replaces/supersedes строится через `tech.NewIndex`. WallSet группируется
  через `civdata.IdentifyWallSets`.

## Команды

```bash
# Сборка
make build                          # → bin/civreport
go build -o civreport ./cmd/civreport

# Один отчёт
make spart                          # → spartans_overview.md + spartans_structree.md
make civ CIV=спарт                  # через алиас
./bin/civreport spart
./bin/civreport --out-dir /tmp spart  # оба файла в /tmp/

# Все цивы (плюс общий common.md)
make all-civs
./bin/civreport --all                # → 15×2 файлов + common.md

# Smoke-тест без записи файлов
make check
./bin/civreport --check

# Указать путь к 0ad
make spart GAMEDATA=/path/to/0ad/binaries/data/mods/public
OAD_GAMEDATA_ROOT=/path ./bin/civreport spart
./bin/civreport --gamedata /path spart

# Конфигурация (перекрывает дефолты из config.json)
./bin/civreport --config ./my-config.json spart
./bin/civreport --lang ru spart                      # gettext-перевод (не реализовано до эпика 5)
./bin/civreport --include-history spart              # история цивы в overview (работает с эпика 2)
./bin/civreport --include-icons spart                # иконки в structree (не реализовано до эпика 5)

# Тесты
make test
go test ./...

# Уборка
make clean                          # удаляет *_overview.md, *_structree.md, common.md, bin/
```

### Опции в `config.json`

Все имеют CLI-эквиваленты. CLI перекрывает JSON, JSON — встроенные дефолты.

```json
{
  "gamedata": "/Users/zeto/Projects/study/0ad/binaries/data/mods/public",
  "lang": "",
  "include_history": false,
  "include_icons": false,
  "out_dir": "."
}
```
