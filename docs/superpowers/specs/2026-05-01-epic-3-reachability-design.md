# Эпик 3 — Reachability, WallSet, Pair, Replaces, Phase Variants

> Дизайн-документ. Принят в брэйнсторминге 2026-05-01.
> Часть дорожной карты из `docs/analysis.md` → «Принятые решения».
> Закрывает gaps **1, 7, 8, 9, 13**.

## Контекст

После эпика 2 overview наполнен, но structree всё ещё опирается на
`filepath.Glob` по `templates/structures/{civ}/` и `templates/units/{civ}/`.
Из-за этого:

- В отчёт попадают «мёртвые» шаблоны (файл в папке цивы есть, но из
  StartEntities недостижим).
- Куски стен (`wall_short`/`wall_medium`/`wall_long`/`wall_gate`/
  `wall_tower`) и обёртка `wallset_stone` рендерятся шестью отдельными
  блоками вместо одного.
- Парные технологии (`pair_*.json`) выводятся одной строкой с пустыми
  ячейками — поля `cost`/`tooltip`/`modifications` живут в `top`/`bottom`,
  но рендер их не разворачивает.
- Цепочки `supersedes`/`replaces`/`replacedBy` храним частично, но
  никак не используем. В колонке «Фаза» все цив-варианты
  (`phase_town_athen`, `phase_town_pers`) сводятся к общему «Town»; для
  Афин и Персов теряются специфичные стоимости и эффекты.

Эпик 3 переводит structree на транзитивное замыкание от
`StartEntities`, группирует WallSet, разворачивает pair-техи и
рендерит цепочки замен/апгрейдов в каждой строке таблицы «Исследует».

## Скоуп

### Включено

- **Транзитивное замыкание от `StartEntities`**. Новый
  `internal/civdata/reach.go` с функцией `Reach(civ, resolver, catalog)`,
  возвращающей `ReachResult{Buildings, Units, Techs, WallSets, Skipped}`.
  Алгоритм fixed-point по отношениям `Trainer/Entities`,
  `Builder/Entities`, `ProductionQueue/Entities`,
  `Trainer/Technologies`, `ProductionQueue/Technologies`,
  `WallSet/Templates`. Подстановка `{civ}`/`{native}` через
  существующий `tmpl.SubstCiv`. *(gap 9)*
- **Удаление `civdata.Buildings()` и `civdata.Units()`**. `Reach` —
  единственный источник истины. Все вызовы (`render/report.go`,
  `numeric_test.go`, `golden_test.go`) мигрируются.
- **Группировка WallSet**. Новый `internal/civdata/wallset.go` с
  `IdentifyWallSets([]Entity) ([]*WallSetGroup, []Entity)`. Возвращает
  список wallset-групп (обёртка + до 8 piece'ов с ролями `Tower`/
  `Gate`/`WallLong`/`WallMedium`/`WallShort`/`Fort`/`WallCurves`/
  `WallEnd`) и список зданий без piece'ов и без обёрток. Wallset
  обёртки и piece'ы из общего `Buildings` исключаются. *(gap 13)*
- **Развёртывание `pair_*.json`**. Новый `internal/tech/pair.go` с
  `ExpandPair(catalog, name) (top, bottom *Technology, ok bool)`.
  Использует уже существующие поля `Pair`/`Top`/`Bottom`. *(gap 1)*
- **Граф `replaces`/`replacedBy`/`supersedes`/`supersededBy`**. Новый
  `internal/tech/replaces.go`: `Index` строится в `NewGenerator`
  через новый `Catalog.LoadAll()` (рекурсивный скан
  `data/technologies/**/*.json`). API: `Index.ResolveForCiv(name, civ)
  *Technology`, `Index.Chain(name) ChainInfo`. *(gap 7, gap 8)*
- **Civ-варианты фаз**. `requirementPhase(req, civ, index)` (новая
  сигнатура) возвращает имя активной для цивы фазы (например, для
  Афин `phase_town` → `phase_town_athen`). В таблице «Исследует»
  отдельная строка на phase-тех показывает civ-вариант с его
  стоимостью и эффектами. *(gap 8)*
- **Рендер цепочек замен**. В таблице «Исследует» под зданием каждая
  строка тех'а получает суффикс «(заменяет: X)» / «(апгрейд от X)» /
  «(заменяется на: Z)» / «(апгрейдится в: Z)» по данным `Index.Chain`.
  Заменённые техи (например, `phase_town`, если у цивы активен
  `phase_town_athen`) скрыты из листинга — единый хелпер
  `resolveTechForCiv` в `render/`.
- **Рендер pair-техов**. `pair_*` в `Trainer/Technologies` здания
  разворачивается в две строки с маркером «◐ парная (выбрать одно)».
  Сам pair-файл в листинге отсутствует. *(gap 1)*
- **Рендер WallSet-блока**. Один заголовок `### Стены: <SpecificName>
  (<GenericName>)`, одна таблица с одной строкой на piece (Tower/Gate/
  WallLong/WallMedium/WallShort плюс опциональные Fort/WallCurves/
  WallEnd). Колонки: Часть, Стоимость, Время, ОЗ, Броня (H/P/C),
  Атака (только у Tower), Длина/прочее. WallSet идёт после обычных
  зданий той же фазы по `BuildingSortKey` от обёртки.
- Тесты:
  - `internal/tech/pair_test.go` — развёртка известного pair (han);
  - `internal/tech/replaces_test.go` — `ResolveForCiv` для phase_town
    у `athen` / `pers` / `germ`;
  - `internal/civdata/reach_test.go` — `spart`: набор зданий
    содержит `civil_centre` и `wonder`, не содержит pieces wallset'а;
  - `internal/civdata/wallset_test.go` — у `spart` ровно две группы
    (palisade + stone), у stone ровно 5 piece'ов;
  - `internal/render/structree_test.go` — наличие секции «Стены»,
    наличие маркера «◐ парная» (для cив с pair-техами).
- `golden_test.go` остаётся структурным (length ≥ 700, наличие
  обязательных секций; список ожидаемых секций обновляется — добавляется
  «Стены»). Untracked golden-файлы регенерируются локально и
  проверяются глазами в рамках приёмки эпика.

### Не включено

- Локализация имён через `.po` (gettext) — эпик 5.
- Иконки в structree — эпик 5.
- `--lang ru` и связанная инфраструктура — эпик 5.
- Skeleton.tmpl миграция — эпик 5/6.
- Новые поля атаки (Bonuses, Splash, Capture, ApplyStatus), Healer-
  таблица, Run speed, Loot, Turrets/Projectiles, Resource trickle —
  эпик 4.
- Авры зданий из `auras/structures/` — эпик 4.

## Архитектура

### Новые/изменённые модули

```
internal/tech/
  ├── pair.go          NEW: ExpandPair(catalog, name) (top, bottom *Tech, ok bool)
  └── replaces.go      NEW: Index, NewIndex(catalog), ResolveForCiv,
                            Chain(name) ChainInfo

internal/civdata/
  ├── enumerate.go     CHANGED: Buildings()/Units()/loadDir() УДАЛЕНЫ.
                            Phase enum, BuildingPhase, BuildingSortKey,
                            GroupByPhase сохраняются.
  ├── reach.go         NEW: Reach(civ, resolver, catalog) (*ReachResult, error)
  └── wallset.go       NEW: IdentifyWallSets([]Entity) ([]*WallSetGroup, []Entity)

internal/render/
  ├── report.go        CHANGED: Generator.Generate использует Reach;
                            Generator получает поле index *tech.Index,
                            строится в New (или лениво в Generate).
  ├── structree.go     CHANGED: renderPhases принимает ReachResult,
                            renderBuilding делегирует на renderWallSetBlock
                            если b — wallset-обёртка. renderResearches
                            учитывает Index. requirementPhase резолвит
                            civ-вариант фазы.
  └── pair.go          NEW: renderPairRows(sb, top, bottom, index, civ)
```

### Алгоритм Reach (fixed-point)

```go
type ReachResult struct {
    Buildings []Entity                    // достижимые structures/.
                                          // Включают wallset-обёртки;
                                          // НЕ включают piece'ы (Tower/
                                          // Gate/WallLong/WallMedium/
                                          // WallShort и др. — попадают
                                          // только в WallSets).
    WallSets  []*WallSetGroup
    Units     []Entity
    Techs     map[string]*tech.Technology // pair_* развёрнут в top+bottom;
                                          // сам pair-файл в map нет.
    Skipped   []SkipNote                  // отсутствующие шаблоны/техи
                                          // для отладочного логирования
}

type SkipNote struct {
    Token  string
    Reason string  // "template not found" | "tech not in catalog" | ...
}

func Reach(civ *Civ, idx *tmpl.Index, resolver *tmpl.Resolver, catalog *tech.Catalog) (*ReachResult, error) {
    seen := map[string]struct{}{}
    queueE := []string{}
    queueT := []string{}
    for _, se := range civ.StartEntities {
        queueE = append(queueE, se.Template)
    }
    res := &ReachResult{Techs: map[string]*tech.Technology{}}
    for len(queueE) > 0 || len(queueT) > 0 {
        for len(queueE) > 0 {
            e := pop(&queueE)
            if e == "" || strings.HasPrefix(e, "-") { continue }
            e = tmpl.SubstCiv(e, civ.Code)
            if _, ok := seen[e]; ok { continue }
            seen[e] = struct{}{}
            path, ok := idx.Lookup(e)              // basename или слэш-путь
            if !ok { res.Skipped = append(res.Skipped, SkipNote{e, "template not found"}); continue }
            el, err := resolver.Resolve(path)
            if err != nil { return nil, err }
            ent := Entity{TemplateID: e, Path: path, Element: el}
            classifyAndAppend(res, ent)
            for _, tok := range el.GetTokens("Trainer/Entities") { queueE = append(queueE, tok) }
            for _, tok := range el.GetTokens("Builder/Entities") { queueE = append(queueE, tok) }
            for _, tok := range el.GetTokens("ProductionQueue/Entities") { queueE = append(queueE, tok) }
            for _, tok := range el.GetTokens("Trainer/Technologies") { queueT = append(queueT, tok) }
            for _, tok := range el.GetTokens("ProductionQueue/Technologies") { queueT = append(queueT, tok) }
            // WallSet/Templates — именованные children (Tower/Gate/WallLong/...),
            // не tokens. Резолвим через el.Get(...).Children.
            if ws := el.Get("WallSet/Templates"); ws != nil {
                for _, child := range ws.Children {
                    if v := strings.TrimSpace(child.Text); v != "" { queueE = append(queueE, v) }
                }
            }
        }
        for len(queueT) > 0 {
            t := pop(&queueT)
            if t == "" || strings.HasPrefix(t, "-") { continue }
            if _, ok := seen["tech:"+t]; ok { continue }
            seen["tech:"+t] = struct{}{}
            techRec, err := catalog.ByName(t)
            if err != nil { res.Skipped = append(res.Skipped, SkipNote{t, "tech not in catalog"}); continue }
            if techRec.Pair != "" || techRec.Top != "" {
                top, bot, ok := ExpandPair(catalog, t)
                if !ok { res.Skipped = append(res.Skipped, SkipNote{t, "pair expansion failed"}); continue }
                res.Techs[top.Name] = top
                res.Techs[bot.Name] = bot
                continue
            }
            res.Techs[t] = techRec
        }
    }
    return res, nil
}
```

Особенности:

- Токены с префиксом `-` уже отфильтрованы резолвером (это удаление из
  tokens-наследования) — но защита остаётся (`strings.HasPrefix(tok, "-")`
  → skip).
- `classifyAndAppend` различает `units/...` и `structures/...` по префиксу
  `TemplateID`, кладёт в `res.Buildings` или `res.Units`.
- `tmpl.Index.Lookup(parentRef)` уже принимает и basename
  (`civil_centre`), и слэш-путь без `.xml` (`units/spart/cavalry_b`) —
  расширять не нужно. Reach получает `*tmpl.Index` отдельным
  параметром (вместо доступа через `Resolver`).
- `Reach` идемпотентен: повторный вызов на той же циве даёт тот же
  результат.

### WallSet-группировка

Запускается **после** `Reach`, на готовом `Buildings`. Алгоритм:

1. Найти все `Entity`, у которых `Element.GetChildren("WallSet/Templates")`
   непустой → это wallset-обёртки.
2. Для каждой обёртки:
   - Прочитать дочерние элементы `WallSet/Templates/*` — их имена и есть
     роли (`Tower`, `Gate`, `WallLong`, ...).
   - Для каждого role-токена найти соответствующий `Entity` в общем
     `Buildings` (по `TemplateID` после `{civ}`-подстановки).
   - Собрать `WallPiece{Role, Entity}`.
3. Собрать `WallSetGroup{Wrapper, Pieces, Phase: BuildingPhase(Wrapper)}`.
4. Из общего `Buildings` исключить и сами обёртки, и все piece'ы.
   Wrapper хранится только в `WallSetGroup`.

```go
type WallSetGroup struct {
    Wrapper Entity
    Pieces  []WallPiece
    Phase   civdata.Phase
}
type WallPiece struct {
    Role   string  // "Tower" / "Gate" / "WallLong" / ...
    Entity Entity
}
```

### Граф замен

```go
type Index struct {
    techs        map[string]*tech.Technology  // basename → tech
    replacedBy   map[string][]string          // X → [Y...] (все техи с replaces ⊇ {X})
    supersededBy map[string]string            // X → Y, если Y.supersedes == X (всегда один)
    replaces     map[string][]string          // зеркало replacedBy
    Warnings     []string                     // diagnostics при сборке
}

// Расширяет tech.Catalog методом LoadAll(), который рекурсивно
// сканирует data/technologies/**/*.json через filepath.WalkDir,
// загружает каждый файл через tech.Load и кладёт в Catalog.cache.
// Index пользуется этим, чтобы не дублировать обход.
func NewIndex(catalog *tech.Catalog) (*Index, error) {
    if err := catalog.LoadAll(); err != nil { return nil, err }
    idx := &Index{techs: catalog.AllLoaded() /* или равнозначный аксессор */}
    for _, t := range idx.techs {
        for _, r := range t.Replaces {
            idx.replacedBy[r] = append(idx.replacedBy[r], t.Name)
            idx.replaces[t.Name] = append(idx.replaces[t.Name], r)
        }
        if t.Supersedes != "" {
            // Если несколько техов имеют один и тот же Supersedes — берём
            // первый стабильно по алфавиту, остальные → idx.Warnings.
            if existing, ok := idx.supersededBy[t.Supersedes]; ok {
                if t.Name < existing { idx.supersededBy[t.Supersedes] = t.Name }
                idx.Warnings = append(idx.Warnings,
                    fmt.Sprintf("multiple supersedes for %q: %s, %s", t.Supersedes, existing, t.Name))
            } else {
                idx.supersededBy[t.Supersedes] = t.Name
            }
        }
    }
    return idx, nil
}

type ChainInfo struct {
    Replaces     []string  // эта тех заменяет (= t.Replaces)
    ReplacedBy   []string  // все техи, заменяющие эту (raw, не civ-aware)
    Supersedes   string    // эта тех апгрейд от ... (= t.Supersedes)
    SupersededBy string    // апгрейдится в ...
}

func (i *Index) Chain(name string) ChainInfo
func (i *Index) ResolveForCiv(name, civ string) *tech.Technology

// Для render: «X заменяется на Y» под конкретной цивой — вызвать
// ResolveForCiv(X, civ). Если результат != idx.techs[X], значит для
// этой цивы X действительно подменяется и в строке нужно показать Y.
```

`ResolveForCiv(name, civ)` алгоритм (точная семантика):

1. Среди `idx.replaces[name]` (все техи с `replaces ⊇ {name}`) выбрать
   кандидата с `requirements.civ == civ`. Если нашёлся — вернуть его.
2. Иначе среди тех же кандидатов выбрать тот, у которого
   `requirements.civ` пуст (generic) и `requirements.notciv` не
   запрещает `civ`. Если ровно один — вернуть его.
3. Если и таких нет — вернуть `idx.techs[name]` (исходный, тоже
   разрешено вернуть `nil`, если он отсутствует).
4. Если на шаге 2 несколько кандидатов — стабильный порядок по
   `Name`, плюс запись в `idx.Warnings`.

Реальный пример: для `name = "phase_town"`, candidates =
`{phase_town_athen, phase_town_pers, phase_town_generic}`. У `germ`
шаг 1 возвращает nil, шаг 2 — `phase_town_generic`. У `athen` —
`phase_town_athen`.

### Поток данных в `Generator.Generate`

```
LoadCiv(civs/{code}.json)
     │
     ├──> NewIndex(catalog)      [один раз для процесса; кешируется в Generator]
     ▼
Reach(civ, resolver, catalog)    ← заменяет Buildings()+Units()
     │
     ▼
res {Buildings, Units, Techs, WallSets}
     │
     ▼
renderHeader / renderOverview / renderStructree(res, index)
     │
     └─ renderPhases:
        ├─ для каждой фазы: обычные здания + WallSetGroups этой фазы
        └─ renderBuilding или renderWallSetBlock
              │
              └─ renderResearches(building, index, civ):
                  for tok := range b.Trainer/Technologies + ProductionQueue/Technologies:
                      if pair: ExpandPair → две строки
                      else: t := index.ResolveForCiv(tok, civ); если t == nil → skip;
                            else → одна строка с suffix из Chain(t.Name)
```

`Generator` получает поле `index *tech.Index`. Конструируется
сразу в `NewGenerator` (там же, где создаётся `Catalog`). Один на
процесс — переиспользуется между всеми вызовами `Generate` (важно
для `--all`, где 15 цив подряд).

### Рендер pair-техов

```markdown
#### Исследует

| Технология | Стоимость | Время | Фаза | Эффект |
|-----------|-----------|-------|------|--------|
| ◐ Civil Service (парная — выбрать одно) | 200 food | 30 сек | Town | +5% к производству ... |
| ◐ Civil Engineering (парная — выбрать одно) | 200 wood | 30 сек | Town | +10% к строительству ... |
| Phase: Kōmopolis (заменяет phase_town) | 500 food, 500 wood | 30 сек | Town | +25% к радиусу влияния CivCentre, ... |
```

Маркер `◐` (U+25D0). Сам pair-файл в листинге отсутствует.

### Рендер WallSet-блока

```markdown
### Стены: Teichos (Stone Walls)

*Каменные стены Спарты. Состоят из пяти типов сегментов; ставятся
после исследования Town Phase.*

| Часть | Стоимость | Время | ОЗ | Броня (H/P/C) | Атака | Прочее |
|-------|-----------|-------|-----|---------------|-------|--------|
| Башня | 100 камень, 50 металл | 8 сек | 2000 | 5/10/0 | 12 (Pierce, range 60) | MaxOverlap 0.9 |
| Ворота | 200 камень | 10 сек | 4000 | 5/10/0 | — | passable |
| Длинная секция | 30 камень | 4 сек | 3000 | 5/10/0 | — | length 18 |
| Средняя секция | 20 камень | 3 сек | 2000 | 5/10/0 | — | length 12 |
| Короткая секция | 10 камень | 2 сек | 1000 | 5/10/0 | — | length 6 |

*Конкретные числа для Спарты — спот-чек в numeric_test.go.*

---
```

Колонка «Прочее» суммирует то, что специфично для роли:
`MaxTowerOverlap` для Tower, `Visible/Footprint/Width` (длина) для
сегментов, признак passable для Gate. Длина — из `Footprint/Square/@width`
в шаблоне.

## Решения и их обоснование

### 1. Reach полностью заменяет Buildings/Units

Альтернатива «Reach как фильтр над Glob» обсуждалась и отвергнута:
оставлять старые функции — значит держать два источника истины и
дополнительный код для синхронизации. Чистый разрез проще.

Ломается только внутренний API проекта. Вызовы — в трёх файлах
(`render/report.go`, `render/numeric_test.go`, `render/golden_test.go`),
все мигрируются в одной задаче.

### 2. Pair-разворот при scan, не при render

Альтернатива «оставить pair'ы в `res.Techs` и разворачивать в
рендере» отвергнута: render не должен лезть в Catalog при каждом
обращении. Развёртка один раз при Reach, в `Techs` лежат уже
top/bottom.

### 3. Index строится сканированием всех techs

Альтернатива «строить ленивые backedges по запросу» отвергнута: full
scan стоит ~150 файлов на старте, проще и предсказуемее. Index можно
кешировать на уровне процесса (полезно для `--all`).

### 4. WallSet — одна таблица без отдельных секций на piece

Решение из брэйнсторминга. Куски функционально различны (Gate
проходим, Tower стреляет, Wall* — пассивные сегменты), а у обёртки
нет собственных боевых параметров — выделять «представителя» некого.
Одна wide-таблица с одной строкой на piece — наиболее честная
раскладка.

### 5. Маркер парного теха `◐`

Unicode «Circle with left half black» (U+25D0). Альтернативы (`⚖`,
`(p)`) обсуждались — `◐` зрительно компактнее и не путается с
другими markdown-символами.

### 6. Generic phase-вариант скрывается, если есть civ-вариант

В `Trainer/Technologies` Civic Centre лежит `phase_town`. У Афин
есть `phase_town_athen` (его `replaces == ["phase_town"]`). Рендер
таблицы «Исследует» под Civic Centre Афин показывает **только**
`phase_town_athen` (с собственной стоимостью и тултипом). Generic
вариант скрыт — иначе пользователь видит две почти одинаковых
строки и не понимает, какая активна.

Для цив без override (Спарта, Германия и др.) показывается generic
`phase_town_generic` (его `replaces == ["phase_town"]` тоже), но
без civ-специфичной части.

### 7. supersedes vs replaces в рендере

`replaces` — «эта тех занимает место X в production-листингах»
(используется и для скрытия, и для суффикса).
`supersedes` — «эта тех является логическим апгрейдом X»
(используется только для суффикса). В чейне фаз: `phase_town_athen`
имеет `replaces=["phase_town"]` и `supersedes="phase_village"` —
оба отрабатывают независимо.

## Тестирование

Все тесты, читающие данные игры, идут через `internal/testutil` и
пропускаются на CI без `OAD_GAMEDATA_ROOT`.

### `internal/tech/pair_test.go`

- `ExpandPair(catalog, "pair_unlock_civil_service_han")` →
  top.Name == "civil_service_01", bot.Name == "civil_service_02".
- `ExpandPair(catalog, "phase_town")` → ok == false.
- Несуществующий pair → ok == false без паники.

### `internal/tech/replaces_test.go`

- `Index.ResolveForCiv("phase_town", "athen")`.Name == `phase_town_athen`.
- `Index.ResolveForCiv("phase_town", "pers")`.Name == `phase_town_pers`.
- `Index.ResolveForCiv("phase_town", "germ")`.Name == `phase_town_generic`
  (для `germ` нет civ-специфичного override, шаг 1 алгоритма даёт nil,
  шаг 2 возвращает уникальный generic-кандидат).
- `Index.Chain("phase_town_athen")`.Replaces == ["phase_town"];
  .Supersedes == "phase_village".
- `Index.Chain("phase_town")`.ReplacedBy содержит как минимум
  `phase_town_generic`, `phase_town_athen`, `phase_town_pers` (полный
  список без civ-фильтрации).

### `internal/civdata/reach_test.go`

- Для `spart`:
  - `len(res.Buildings) >= 10`.
  - `civil_centre`, `wonder`, `barracks`, `forge` присутствуют.
  - `wall_short`/`wall_medium`/`wall_long`/`wall_gate`/`wall_tower`
    отсутствуют в `Buildings` (они в WallSets).
  - `wallset_stone` отсутствует в `Buildings` (он в WallSets).
  - `len(res.Units) >= 5`, включает `support_civilian`,
    `infantry_spearman_b`.
  - `res.Techs` содержит `phase_town`-чейн.
- Идемпотентность: повторный вызов даёт ту же длину Buildings/Units.

### `internal/civdata/wallset_test.go`

- Для `spart`: ровно 2 группы (palisade — generic structures/wallset_palisade
  и stone — structures/spart/wallset_stone). У stone ровно 5 piece'ов
  (Tower/Gate/WallLong/WallMedium/WallShort). У piece'ов basename
  совпадает с `wall_tower`/`wall_gate`/`wall_long`/`wall_medium`/
  `wall_short`.
- Для `pers`: палисад + stone. Spec-проверка Persia опциональна.

### `internal/render/structree_test.go`

- Структурная проверка: после генерации structree для `germ` строка
  «### Стены» присутствует ровно 2 раза (palisade + stone).
- Для civ с pair-техом (например `han`): маркер «◐ парная» присутствует
  хотя бы 1 раз.
- Для `athen`: строка «phase_town_athen» или «Kōmopolis» присутствует;
  строка «phase_town_generic» отсутствует.

### `internal/render/numeric_test.go`

- Перевод с `Buildings(...)` на `Reach(...).Buildings`. Тесты на
  конкретные числа (HP/броня spart spearman) сохраняются как есть.

### `internal/render/golden_test.go`

- Список ожидаемых секций обновляется (добавляется «Стены»).
- Длина ≥ 700 строк сохраняется.
- Untracked golden-файлы регенерируются локально и проверяются глазами
  в рамках приёмки эпика. Никаких byte-diff проверок.

## План коммитов (для subagent-driven dev)

Согласно памяти `feedback_workflow_subagent_driven` — пары «main + nits»:

1. **tech: pair expansion** — `pair.go` + `pair_test.go`. Nits: gofmt,
   doc-comment.
2. **tech: replaces/supersedes index** — `replaces.go` + `replaces_test.go`.
   Nits: comment accuracy, Empty-Index edge cases.
3. **civdata: Reach algorithm** — `reach.go` + `reach_test.go`. Nits.
4. **civdata: WallSet grouping** — `wallset.go` + `wallset_test.go`.
   Удаление `Buildings()`/`Units()`/`loadDir()` в этом же коммите.
   Nits.
5. **render: integrate Reach + Index** — `report.go`, `structree.go`,
   `pair.go` (рендер), `numeric_test.go`, `golden_test.go`,
   `structree_test.go`. Nits.
6. **docs: sync analysis.md/current-state.md/CLAUDE.md под эпик 3**.

После 5 — регенерация untracked goldens (`make germ`) и code-reviewer
субагент.

## Открытые вопросы

Ни одного — все решения зафиксированы в брэйнсторминге.
