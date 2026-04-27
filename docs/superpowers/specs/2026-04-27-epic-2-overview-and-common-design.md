# Эпик 2 — Overview + common.md

> Дизайн-документ. Принят в брэйнсторминге 2026-04-27.
> Часть дорожной карты из `docs/analysis.md` → «Принятые решения».
> Закрывает gaps **2, 3, 4, 5, 6, 20, 21, 22**.

## Контекст

После эпика 1 структура вывода переключена на пару `<civ>_overview.md` +
`<civ>_structree.md` плюс общий `common.md`, но содержательно overview
тонкий (4 секции из старого `report.go`), а `common.md` — заглушка.
Эпик 2 наполняет обе вкладки данными, эквивалентными игровой
Civilization Overview, и собирает в `common.md` глобальную справку,
на которую цив-отчёты будут ссылаться вместо повторений.

## Скоуп

### Включено

- Загрузка Player-шаблона `templates/special/players/<civ>.xml` через
  существующий `tmpl.Resolver`. Поля `Identity/GenericName`,
  `Identity/History`, `Identity/Icon`. Опция `IncludeHistory`
  активируется в overview. *(gap 20)*
- Одиночный загрузчик team-bonus авры из
  `data/auras/teambonuses/<civ>_player_teambonus.json`. Структурированный
  блок «Командный бонус» в overview с `auraName`, `auraDescription`,
  таблицей `modifications`. *(gap 2)*
- Сканер глобальных autoResearch-техов из корня
  `data/technologies/`. Полное описание `unit_advanced`/`unit_elite`
  переезжает в `common.md`; в overview одна сноска со ссылкой на якоря.
  *(gap 3)*
- Парсер per-modification `affects`. Поле `Modification.AffectsRaw`,
  геттер `AffectsList() []string`. `i18n.DescribeModification` дописывает
  суффикс `(только {affects})`. *(gap 6)*
- Глубокий разбор `requirements`: entity/all/any/notciv/civ/tech.
  `tech.RequirementsTooltip` парсится как fallback-источник. Новая
  колонка «Требования» в таблице Civ Bonuses. *(gap 4 — общая
  инфраструктура; колонка «Фаза» в structree остаётся на эпик 4)*
- Lookup `specificName[civ]` для имени теха в Civ Bonuses таблице.
  *(gap 5)*
- Секция «Уникальные строения» в overview по фильтру класса
  `CivSpecific Structure`. *(gap 21, structures)*
- Секция «Уникальные технологии» в overview — короткий список из
  `Catalog.AllCivBonuses(civ)`. *(gap 21, technologies)*
- Секция «Герои» в overview — короткая выборка имени/класса/одной
  строки описания первой найденной ауры. *(gap 22)*
- Наполнение `common.md`:
  - модификаторы Advanced/Elite целиком из
    `unit_advanced.json`/`unit_elite.json`;
  - типы урона из `data/damage_types/*.json`;
  - типы ресурсов из `data/resources/*.json` (с подтаблицей subtypes);
  - статус-эффекты из `data/status_effects/*.json`.
- CLI: новый флаг `--include-history`, привязка к
  `Config.IncludeHistory`.
- Тесты: unit-тесты на каждый новый загрузчик/форматтер;
  обновление структурного `golden_test.go` под новые секции.

### Не включено

- Дефолты `template_unit.xml`/`template_structure.xml` для `common.md`
  (отложено на эпик 4 — не имеет ценности для overview, требует
  отдельного анализа).
- WallSet-группировка, transitive closure, pair-tech expansion,
  building auras, attack bonuses/splash/capture, healer/turrets,
  run speed, capture-resistance, supersedes/replaces consumption —
  эпики 3–4.
- Локализация `.po` (`--lang`) и иконки в structree (`--include-icons`)
  — эпик 5.
- Колонка «Фаза» в structree-таблице `Researches` — остаётся как есть,
  глубокая переработка → эпик 4.
- Новые блоки в `<civ>_structree.md`. Structree получает изменения
  per-mod affects и requirementsTooltip как **побочный эффект** общих
  функций i18n/tech, без правки structree-кода.

## Решения, принятые в брэйнсторминге

1. **Tech-модель улучшаем как общую инфраструктуру** (вариант A в
   вопросе скоупа). `Modification.Affects`, `RequirementsTooltip`,
   `GlobalAutoResearch`, `DescribeRequirements`, lookup specificName —
   меняются в shared-модулях, structree подхватывает «бесплатно».
2. **История цивы — opt-in через `--include-history`,** дефолт false.
   Соответствует решению 4 в `analysis.md`.
3. **Глобальные autoResearch-техи описываются в `common.md`,**
   overview-секция Civ Bonuses содержит только сноску со ссылкой на
   якоря `common.md`.
4. **`common.md` в этом эпике закрывает 4 раздела** (Advanced, Elite,
   damage types, resources, status_effects). Дефолты template-шаблонов
   — позже.
5. **Cross-references из overview → common.md** добавляются в этом эпике
   (одна footer-сноска в Civ Bonuses секции).
6. **«Уникальные технологии» как короткий список,** отдельно от
   таблицы Civ Bonuses (где даны эффекты с требованиями). Это два разных
   ракурса на одно множество техов; дублирования имён в выводе мы
   осознанно допускаем — короткий список даёт быстрый обзор «чем
   уникальна цива», таблица — детали.
7. **Heroes-блок краткий**: имя + class + 1 строка `auraDescription`
   первой ауры. Полная таблица аур остаётся в Appendix
   `<civ>_structree.md` (renderUnitsDetail).
8. **Командный бонус — структурированная таблица модификаций,** не
   только текстовое описание. Имя из `civ.json.TeamBonuses[0].Name`,
   `auraName`/`auraDescription` — из team-bonus JSON.
9. **CivSpecific фильтр для structures — по классам сущности**
   (`Identity/Classes` token-список содержит "CivSpecific" и
   "Structure").
10. **План коммитов — 7 коммитов в одном PR.** Каждый компилируется,
    `go test ./...` проходит. Subagent-driven development применим.

## Архитектура

### Поток данных

```
Config.IncludeHistory ─┐
                       │
LoadCiv(json) ─────────┤
LoadPlayerTemplate ────┤
LoadTeamBonus ─────────┤
AllCivBonuses ─────────┼─→ renderOverview → Output.Overview
heroAuras ─────────────┤      (footer-link → common.md anchors)
buildings (filter
 CivSpecific) ─────────┤
AllNotCiv ─────────────┘

LoadTech("unit_advanced") ─┐
LoadTech("unit_elite") ────┤
DamageTypes() ─────────────┼─→ renderCommon → common.md
Resources() ───────────────┤
StatusEffects() ───────────┘
```

### Целевая структура пакетов

```
cmd/civreport/main.go              + биндинг --include-history
internal/config/config.go          без изменений (поле уже есть)
internal/paths/paths.go            + PlayerTemplate, TeamBonus,
                                     DamageTypes, Resources, StatusEffects
internal/tmpl/                     без изменений
internal/civdata/
  ├── codes.go                     без изменений
  ├── civ.go                       без изменений
  ├── enumerate.go                 без изменений
  ├── player.go                    NEW: LoadPlayerTemplate(layout, civ)
  └── player_test.go               NEW
internal/aura/
  ├── aura.go                      без изменений
  ├── teambonus.go                 NEW: LoadTeamBonus(layout, civ)
  └── teambonus_test.go            NEW
internal/tech/
  ├── tech.go                      + AffectsRaw на Modification,
                                     RequirementsTooltip, Replaces[]
  ├── tech_test.go                 NEW (Affects парсинг)
  ├── auto.go                      NEW: Catalog.GlobalAutoResearch()
  ├── auto_test.go                 NEW
  ├── requirements.go              NEW: DescribeRequirements()
  └── requirements_test.go         NEW
internal/i18n/
  ├── ru.go                        без существенных изменений
  ├── modifier.go                  + per-mod affects суффикс,
                                     TechDisplayName(specificName lookup)
  └── modifier_test.go             NEW
internal/render/
  ├── skeleton/                    без изменений
  ├── overview.go                  REWRITE: новые секции 1–9
  ├── overview_test.go             NEW
  ├── common.go                    REWRITE: 6 секций контента
  ├── common_test.go               NEW
  ├── structree.go                 без изменений
  ├── format.go                    без изменений
  ├── report.go                    + поле Generator.IncludeHistory,
                                     LoadPlayerTemplate + LoadTeamBonus
                                     в Generate()
  ├── numeric_test.go              без изменений
  └── golden_test.go               + проверка новых заголовков
testdata/golden/
  ├── germans_overview.md          обновляется (untracked, не коммитится)
  └── germans_structree.md         без изменений
```

### Изменения в `internal/civdata/`

```go
// player.go (NEW)
package civdata

import (
    "github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
    "github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type PlayerTemplate struct {
    GenericName string
    History     string
    IconPath    string
}

func LoadPlayerTemplate(layout paths.Layout, civCode string,
    resolver *tmpl.Resolver) (*PlayerTemplate, error) {
    el, err := resolver.Resolve(layout.PlayerTemplate(civCode))
    if err != nil {
        return nil, err
    }
    return &PlayerTemplate{
        GenericName: el.GetText("Identity/GenericName"),
        History:     unescapeNewlines(el.GetText("Identity/History")),
        IconPath:    el.GetText("Identity/Icon"),
    }, nil
}

// unescapeNewlines заменяет литеральные \n в History на реальные
// абзацы для markdown.
func unescapeNewlines(s string) string {
    return strings.ReplaceAll(s, `\n`, "\n\n")
}
```

### Изменения в `internal/aura/`

```go
// teambonus.go (NEW)
package aura

import (
    "github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func LoadTeamBonus(layout paths.Layout, civCode string) (*Aura, error) {
    return Load(layout.TeamBonus(civCode))
}
```

### Изменения в `internal/tech/`

```go
// tech.go: Modification + Technology
type Modification struct {
    Value      string          `json:"value"`
    Multiply   float64         `json:"multiply,omitempty"`
    Add        float64         `json:"add,omitempty"`
    Replace    any             `json:"replace,omitempty"`
    AffectsRaw json.RawMessage `json:"affects,omitempty"` // string or []string
}

func (m Modification) AffectsList() []string {
    if len(m.AffectsRaw) == 0 {
        return nil
    }
    var s string
    if err := json.Unmarshal(m.AffectsRaw, &s); err == nil {
        return []string{s}
    }
    var arr []string
    if err := json.Unmarshal(m.AffectsRaw, &arr); err == nil {
        return arr
    }
    return nil
}

type Technology struct {
    // ... существующие поля
    RequirementsTooltip string   `json:"requirementsTooltip"`
    Replaces            []string `json:"replaces"`
}
```

```go
// auto.go (NEW)
package tech

// GlobalAutoResearch возвращает техи из корня technologies/, у которых
// autoResearch=true и нет requirements.civ (= применяются всем цивам
// автоматически). Сюда попадают unit_advanced, unit_elite, phase_village,
// soldier_ranged_experience, upgrade_rank_advanced_mercenary (R28).
//
// Civ-specific автотехи (`civbonuses/<civ>_*.json`) отфильтрованы: они
// уже доступны через AllCivBonuses(civ).
func (c *Catalog) GlobalAutoResearch() ([]*Technology, error) {
    out := []*Technology{}
    matches, err := filepath.Glob(filepath.Join(c.dir, "*.json"))
    if err != nil {
        return nil, err
    }
    sort.Strings(matches)
    for _, p := range matches {
        t, err := Load(p)
        if err != nil {
            return nil, err
        }
        if !t.AutoResearch {
            continue
        }
        if RequiresCiv(t.Requirements) != "" {
            continue
        }
        out = append(out, t)
    }
    return out, nil
}
```

```go
// requirements.go (NEW)
package tech

// DescribeRequirements возвращает человекочитаемое описание requirements
// (entity/all/any/notciv/civ/tech) на русском. Используется как
// fallback к Technology.RequirementsTooltip.
func DescribeRequirements(req Requirements) string { /* ... */ }
```

### Изменения в `internal/i18n/`

```go
// modifier.go
func DescribeModification(m tech.Modification) string {
    target := translatePath(m.Value)
    var body string
    switch {
    case m.Multiply != 0:
        body = fmt.Sprintf("%s %s", target, FormatPercent(m.Multiply))
    case m.Add != 0:
        // ... как сейчас
    case m.Replace != nil:
        body = fmt.Sprintf("%s = %v", target, m.Replace)
    default:
        body = target
    }
    if affects := m.AffectsList(); len(affects) > 0 {
        body += fmt.Sprintf(" (только %s)", strings.Join(affects, "+"))
    }
    return body
}

// TechDisplayName возвращает отображаемое имя теха для текущей цивы.
// Если specificName[civ] определён — добавляет «(локально: <name>)».
func TechDisplayName(t *tech.Technology, civ string) string {
    name := t.GenericName
    if name == "" {
        name = t.Name
    }
    if t.SpecificName == nil {
        return name
    }
    if v, ok := t.SpecificName[civ]; ok {
        if s, ok := v.(string); ok && s != "" {
            return fmt.Sprintf("%s (локально: %s)", name, s)
        }
    }
    return name
}
```

### Изменения в `internal/paths/paths.go`

```go
func (l Layout) PlayerTemplate(civ string) string {
    return filepath.Join(l.Templates(), "special", "players", civ+".xml")
}
func (l Layout) TeamBonus(civ string) string {
    return filepath.Join(l.Auras(), "teambonuses", civ+"_player_teambonus.json")
}
func (l Layout) DamageTypes() string {
    return filepath.Join(l.Root, "simulation", "data", "damage_types")
}
func (l Layout) Resources() string {
    return filepath.Join(l.Root, "simulation", "data", "resources")
}
func (l Layout) StatusEffects() string {
    return filepath.Join(l.Root, "simulation", "data", "status_effects")
}
```

### Структура `<civ>_overview.md` после эпика 2

```markdown
# {Имя} ({CIV}) — Civilization Overview
> info-блок (Date, IncludeHistory, IncludeIcons, Lang)
---

## Идентичность
- **Код:** `spart`
- **Культура:** hellenic
- **Имя в данных:** Spartans
- **Эмблема:** `emblems/emblem_spartans.png`
- **Стартовые юниты:** structures/spart/civil_centre ×1, …

## Историческая справка    ← только при --include-history
> Sparta was a prominent city-state in Ancient Greece…

## Герои
- **Leonidas I** — Hero, Infantry, Champion. *Inspires nearby Spartan…*
- **Brasidas** — Hero, Infantry. *…*
- **Agis II** — Hero, Cavalry. *…*

## Уникальные строения
- **Syssiton** — *Train heroes here.*
- **Theater** — *Allows training of …*

## Уникальные технологии
- **Peloponnesian League** — *…*
- **Spartan Soldiers** — *…*
- (и т.д., из AllCivBonuses)

## Цивилизационные бонусы
| Бонус | Источник | Требования | Эффект |
|-------|----------|------------|--------|
| Spartan Soldiers (локально: …) | civbonuses/… | civ=spart | рубящий урон ближнего боя +10% (только Melee) |
| … | civ JSON | — | … |

> Глобальные авто-эффекты при повышении ранга применяются ко всем
> цивам — см. [common.md#модификаторы-advanced](common.md#модификаторы-advanced)
> и [#модификаторы-elite](common.md#модификаторы-elite).

## Командный бонус
**Peloponnesian League.** Heroes are trained for free.
- Тип ауры: `global`
- Цель: `Hero`
- Игроки: `MutualAlly`

| Цель | Эффект |
|------|--------|
| стоимость еды | = 0 |
| стоимость дерева | = 0 |
| стоимость камня | = 0 |
| стоимость металла | = 0 |

## Технологии, недоступные spartans
- **Hellenistic Metropolis** — …
- …

---
```

### Структура `common.md` после эпика 2

```markdown
# Common Reference (всеобщие данные для всех цив)
> info-блок
---

## Модификаторы Advanced
> Advanced and Elite units +20% training time, +25% health, …

| Цель | Эффект |
|------|--------|
| захват | +0.7 |
| рубящий урон ближнего боя | +10% (только Melee) |
| рубящий урон ближнего боя | +10% (только Melee) |
| … | … |

## Модификаторы Elite
…

## Прочие глобальные авто-эффекты

Сюда попадают autoResearch-техи без `requirements.civ`, не относящиеся
к ранговому апгрейду:

- **Village Phase** (phase_village) — фазовый авто-тех с локальными
  именами по цивам (Kōmē, Vicus, Ialon, ...).
- **Ranged Soldiers experience** (soldier_ranged_experience) — +25%
  опыта ранга для дистанционных солдат.
- **Expertise In War** (upgrade_rank_advanced_mercenary) —
  наёмники начинают на ранге Advanced.

| Имя | Эффект |
|-----|--------|
| Village Phase | (technology unlocking village-tier buildings) |
| Ranged Soldiers experience | требуемый опыт ранга +25% (только Ranged Soldier) |
| Expertise In War | требуемый опыт ранга = 0 (только Mercenary Basic) |

## Типы урона
| Код | Имя | Описание |
|-----|-----|----------|
| Hack | Hack | Damage caused by sharp objects… |
| Pierce | Pierce | … |
| Crush | Crush | … |
| Fire | Fire | … |
| Poison | Poison | … |

## Типы ресурсов
| Код | Имя | Описание |
|-----|-----|----------|
| food | Food | Harvest from animals, berry bushes… |
| wood | Wood | … |
| stone | Stone | … |
| metal | Metal | … |

#### Подтипы ресурсов
- **food**: fish, fruit, grain, meat
- **wood**: tree, ruins
- …

## Статус-эффекты
| Код | Имя | На применяющего | На пострадавшего |
|-----|-----|-----------------|------------------|
| Burning | Burning | This unit causes fire damage. | This unit is burning. |
| Poisoned | Poisoned | … | … |
```

## Стратегия миграции

Эпик идёт 7 коммитами в одном PR (`epic-2-overview-common`). Каждый
компилируется, `go test ./...` проходит. Применим
subagent-driven-development (Sonnet × 3 ролями), как зафиксировано
в memory.

### Коммит 1 — `civdata: load Identity from special/players/<civ>.xml`

Создаёт `internal/civdata/player.go` с `LoadPlayerTemplate` и
`internal/civdata/player_test.go`. Зависимостей нет — никто не зовёт.
Тесты:
- `TestLoadPlayerTemplate_Spart` — резолвит `spart.xml`,
  GenericName="Spartans", History начинается с "Sparta was a prominent
  city-state", IconPath="emblems/emblem_spartans.png".
- Тест skip'ается, если корень данных недоступен (паттерн как в
  существующих numeric/resolver тестах).

### Коммит 2 — `aura: add LoadTeamBonus single-file loader`

Создаёт `internal/aura/teambonus.go` с `LoadTeamBonus` (тонкая
обёртка над `aura.Load`). Тест: `TestLoadTeamBonus_Spart` —
AuraName="Peloponnesian League", `Modifications` содержит
`Cost/Resources/food → replace 0`.

### Коммит 3 — `tech: per-mod affects, RequirementsTooltip, GlobalAutoResearch, DescribeRequirements`

- Расширяет `Modification` полем `AffectsRaw json.RawMessage` + метод
  `AffectsList()`.
- Добавляет `Technology.RequirementsTooltip` и `Technology.Replaces`.
- Создаёт `internal/tech/auto.go` с `Catalog.GlobalAutoResearch()`.
- Создаёт `internal/tech/requirements.go` с `DescribeRequirements`.
- Тесты: `tech_test.go` (Affects парсинг string и []string),
  `auto_test.go` (содержит unit_advanced+unit_elite, не содержит
  spart_population/phase_village/soldier_ranged_experience),
  `requirements_test.go` (entity/all/any/notciv/civ/tech).

`i18n.DescribeModification` пока не зовёт `AffectsList()` — это
коммит 4.

### Коммит 4 — `i18n: per-mod affects suffix; TechDisplayName(specificName)`

- `DescribeModification` дописывает суффикс `(только {affects})`.
- Новая `TechDisplayName(t, civ)` с specificName-лукапом.
- Тесты: `modifier_test.go`.

К концу коммита structree-таблица `Researches` уже выводит per-mod
affects суффиксы (без дополнительной правки structree.go).

### Коммит 5 — `render: rewrite renderOverview with epic-2 sections`

- `internal/render/overview.go` переписан полностью.
- `Generator` получает дополнительное поле `IncludeHistory bool`
  (рендер не нуждается в полной `*config.Config` — это лишний coupling).
- `Generator.Generate` подгружает `LoadPlayerTemplate` и
  `LoadTeamBonus`, передаёт в renderOverview.
- `paths.Layout` получает `PlayerTemplate(civ)`/`TeamBonus(civ)`.
- **Все секции overview на уровне `## `** (подтаблицы внутри секции
  — `### `). Это меняет уровень заголовков относительно эпика 1
  (`### Цивилизационные бонусы` → `## Цивилизационные бонусы`),
  поэтому golden_test обновляется в коммите 7.
- Тесты: `overview_test.go` (структурный — наличие секций,
  подстроки `Peloponnesian League`, `Spartans`).

### Коммит 6 — `render: populate common.md`

- `internal/render/common.go` переписан: 6 секций.
- `paths.Layout` получает `DamageTypes()`/`Resources()`/`StatusEffects()`.
- Загрузка `unit_advanced`/`unit_elite` через `Catalog.ByName`.
- Загрузка damage/resources/status — собственные минимальные
  json-парсеры (struct + `os.ReadDir` + сортировка по `order` для
  damage и resources).
- Тесты: `common_test.go` (наличие 6 секций, проверка содержимого
  Advanced на наличие `Health/Max` и `multiply 1.25`).

### Коммит 7 — `cmd,config: bind --include-history; update docs and golden_test`

- `cmd/civreport/main.go`: добавляет `--include-history`,
  пробрасывает в Generator.
- `internal/render/golden_test.go`:
  - **удаляет** старую проверку `common` на подстроку «TODO» (epic-1
    плейсхолдер);
  - **удаляет** старые проверки заголовков `## Общая информация о
    цивилизации`, `### Цивилизационные бонусы`;
  - **добавляет** проверки заголовков `## Идентичность`, `## Герои`,
    `## Уникальные строения`, `## Уникальные технологии`,
    `## Цивилизационные бонусы`, `## Командный бонус`,
    `## Технологии, недоступные …`;
  - **добавляет** проверки common-секций `## Модификаторы Advanced`,
    `## Модификаторы Elite`, `## Типы урона`, `## Типы ресурсов`,
    `## Статус-эффекты`;
  - порог `overviewLines` поднимается с 25 до 50.
- README.md и CLAUDE.md обновляются: новый флаг, новые секции overview,
  наполнение common.md.

## Тесты

| Файл | Что проверяет |
|------|----------------|
| `internal/civdata/player_test.go` | spart.xml: GenericName/History/Icon после резолва |
| `internal/aura/teambonus_test.go` | spart team-bonus AuraName и modifications |
| `internal/tech/tech_test.go` | Modification.Affects: string-форма и []string-форма |
| `internal/tech/auto_test.go` | GlobalAutoResearch включает `unit_advanced`, `unit_elite`, `phase_village`, `soldier_ranged_experience`, `upgrade_rank_advanced_mercenary` (все autoResearch без `requirements.civ`); НЕ включает civ-specific бонусы из `civbonuses/` (они отфильтрованы по `RequiresCiv()!=""`) |
| `internal/tech/requirements_test.go` | entity{class,number}, all-mix, any, notciv → русская строка |
| `internal/i18n/modifier_test.go` | суффикс `(только Melee)`; TechDisplayName с/без specificName |
| `internal/render/overview_test.go` | для spart: содержит 8 заголовков; команд-бонус содержит «Peloponnesian League»; история отсутствует без флага и присутствует с флагом |
| `internal/render/common_test.go` | 6 секций, конкретные значения unit_advanced |
| `internal/render/golden_test.go` | расширенный список заголовков для germ_overview |
| `internal/render/numeric_test.go` | без изменений |

## Definition of Done

1. На ветке `epic-2-overview-common` 7 коммитов; каждый компилируется
   и `go test ./...` проходит.
2. `civreport spart` пишет `spartans_overview.md`,
   `spartans_structree.md`, `common.md`. Overview содержит 8
   обязательных секций.
3. `civreport --include-history spart` дополнительно содержит блок
   «Историческая справка».
4. `common.md` содержит 6 секций (Модификаторы Advanced, Модификаторы
   Elite, Прочие глобальные авто-эффекты, Типы урона, Типы ресурсов,
   Статус-эффекты).
5. `civreport --all` пишет 30 цив-файлов + общий `common.md`.
6. `civreport --check` зелёный для всех 15 цив.
7. Все unit/golden тесты зелёные.
8. README + CLAUDE.md актуализированы (флаг `--include-history`, новые
   секции overview, наполнение common.md).
9. В overview Civ Bonuses таблица содержит колонку «Требования», а
   модификации с `affects` показываются с суффиксом `(только …)`.
10. Якорные ссылки `common.md#модификаторы-advanced` /
    `#модификаторы-elite` рендерятся в каждом overview-файле.

## Risks and mitigations

- **`Modification.affects` JSON: string vs []string.** Использован
  паттерн `RawMessage` + геттер (как `Civ.Culture` и `Aura.Affects`).
  Тест на обе формы.
- **`Identity/History` хранит литеральные `\n`.** Парсим как есть,
  заменяем `\\n` → `\n\n` на стороне `LoadPlayerTemplate` для
  markdown-параграфов. Не трогаем shared template-логику.
- **`GlobalAutoResearch` может зацепить нежелательное.**
  Двух-этапный фильтр (`autoResearch=true` И `requires.civ==""` И нет
  `req.tech`). Для R28 это даёт ровно `unit_advanced` + `unit_elite`.
  Тест проверяет точный список.
- **CivSpecific structures могут быть пустыми у части цив.**
  Секция показывается всегда; при пустом списке — markdown-плейсхолдер
  «*У цивы нет уникальных строений.*».
- **Order файлов для overview.** Использовать алфавитную сортировку по
  basename (детерминизм; матчит требование CLAUDE.md «Все списки
  сортируются перед рендером»).
- **Anchor-генерация markdown-движков.** Для якоря `common.md#модификаторы-advanced`
  предполагаем GitHub-style slug-генерацию (lowercase, дефисы вместо
  пробелов, кириллица сохраняется). Если потребитель использует другой
  движок — anchor может не сработать. Митигация: оставить ссылку как
  есть, отметить в README известный нюанс. Альтернатива (`<a name=...>`-
  якори) — сложнее, не обязательно для CLI-формата.
- **Increase в overview длине.** Текущий overview ~30 строк; после
  эпика 2 — ~80–150 строк. Smoke-порог `--check` для overview можно
  поднять до 50 строк (был 30).

## Open questions

Нет — все вопросы либо зафиксированы решениями выше, либо вынесены в
будущие эпики (3, 4, 5).
