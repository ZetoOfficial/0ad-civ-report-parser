# Эпик 4a — Attack & Resistance details

> Дизайн-документ. Принят в брэйнсторминге 2026-05-02.
> Часть дорожной карты из `docs/analysis.md` → «Принятые решения».
> Закрывает gaps **10, 11, 17**.

## Контекст

После эпика 3 структура отчёта стабилизирована: `Reach`-граф,
WallSet-группировка, разворот пар-техов и явные цепочки `replaces`.
Всё, что относится к **подробностям атаки и сопротивления юнитов**,
до сих пор сводится к двум плоским строкам в `renderUnitBlock`
(`Атака (ближ.)`, `Атака (стрельба)`) и одной строке `Броня (H/P/C)`.

Это даёт частичную картину: champion Pikeman без явного `×2.5 vs Cavalry`
читается как «обычный пехотинец»; siege stonethrower без splash-блока
выглядит как лёгкий лучник; индийский слон с Poisoned-эффектом не
показывает свойство, ради которого его нанимают.

Эпик 4a добавляет в `renderUnitBlock` (и частично в `renderBuilding`)
все недостающие детали атаки и сопротивления, оставляя плоскую
2-колоночную таблицу как несущую конструкцию (вариант A,
зафиксирован в брэйнсторминге).

Эпик 4b (отдельный документ — будет писаться по завершении 4a)
закрывает оставшиеся gaps 12, 14, 15, 16, 18: ауры зданий, Loot/
Trickle/Dropsite/Repair, Healer-таблица, Turrets/Projectiles,
Run speed.

## Скоуп

### Включено

- **Bonuses в атаке** *(gap 10)* — `<Bonuses>` в `<Melee>` / `<Ranged>` /
  `<Capture>` развёртывается в отдельную строку
  `Бонусы (<режим>)` со значением вида `×2.5 vs Cavalry` (несколько
  бонусов через `, `; несколько классов внутри одного бонуса через
  `+`, как `×3 vs Cavalry+Elephant`).
- **PreferredClasses** *(часть gap 10)* — `<PreferredClasses>` в каждом
  attack-mode развёртывается в отдельную строку
  `Предпочитает (<режим>)`. Семантика — приоритет цели в UnitAI, не
  damage-модификатор (`getAttackBonusTooltip` отделяет в pyrogenesis).
- **Splash в атаке** *(часть gap 11)* — `<Splash>` в attack-mode
  развёртывается в строку `Брызги (<режим>)` со значением вида
  `Crush 120, круг R=1.5, не задевает союзников`.
- **Capture как режим атаки** *(часть gap 11)* — `<Attack>/<Capture>`
  развёртывается в отдельные строки параллельно `Атака (ближ.)` /
  `Атака (стрельба)`: `Атака (захват) | захват 2.5 (4м, 1000мс);
  исключает: Field, Palisade, Wall`. Всегда печатается, если
  `Attack/Capture` существует — в т. ч. при дефолте, наследуемом из
  `template_unit_*.xml`.
- **ApplyStatus в атаке** *(часть gap 11, locked decision №8)* —
  для каждого статус-кода внутри `<ApplyStatus>` пишется одна строка
  `Накладывает (<режим>) | Poisoned: Poison 2/0.6с × 3.6с (см.
  common.md#poisoned)`. Поддерживается переменный `Interval` и
  `Duration` (нормализуем к секундам). Если у статус-эффекта поле
  `BlockChance > 0` и `Duration == 0` — это «no-op default» из базового
  шаблона, такие узлы пропускаем (см. § «Edge case: BlockChance»).
- **Resistance/Entity/Capture** *(часть gap 17)* — строка
  `Сопротивление захвату | <число>` после `Броня (H/P/C)` в
  unit-блоке. У зданий — та же строка в `renderBuilding`.
- **Resistance/Entity/StatusEffect** *(часть gap 17)* — строка
  `Сопротивление статус-эффектам | poisoned ×0.5, burning ×0.7`
  после строки сопротивления захвату. Каждый ребёнок —
  `<<code>>=<multiplier>`; вывод в lowercase, чтобы не путать с
  CamelCase именами в attack-блоке.
- **Под-якоря в common.md для статус-эффектов** — `commonStatusEffects`
  в дополнение к существующей таблице пишет короткие per-status
  заголовки `### Poisoned` / `### Burning` (с `applierTooltip` и
  `receiverTooltip`), чтобы GitHub-anchors `#poisoned`, `#burning`
  существовали и cross-file ссылки из unit-строк работали.
- **Расширение `i18n.DamageType`** — добавляются ключи `Fire`, `Poison`
  (используются в `<Damage>` внутри ApplyStatus). Существующие
  Hack/Pierce/Crush сохраняются.
- Тесты:
  - `internal/render/format_test.go` — pure-функции на inline-XML
    фикстурах для `formatAttackBonuses`, `formatPreferredClasses`,
    `formatSplash`, `formatCaptureAttack`, `formatApplyStatuses`,
    `FormatCaptureResistance`, `FormatStatusEffectResistance`.
    Edge cases: пустой `<Bonuses>`, отсутствующий `FriendlyFire`,
    `Interval != 1000`, BlockChance-default, мульти-status.
  - `internal/render/structree_test.go` — структурные grep-проверки
    на готовом markdown:
    - для champion pikeman есть `| Бонусы (ближ.) | ×2.5 vs Cavalry |`;
    - для siege stonethrower есть `| Брызги (стрельба) |` и
      `Crush 120, круг R=1.5`;
    - для maur champion maiden archer есть `| Накладывает (стрельба) |
      Poisoned: Poison 2/0.6с × 3.6с (см. common.md#poisoned) |`;
    - для иберийского champion cavalry — Burning-строка по аналогии;
    - для любого пехотинца — `| Атака (захват) | захват 2.5 (4м,
      1000мс); исключает: Field, Palisade, Wall |`.
  - `internal/render/numeric_test.go` — спот-чек: спартанский champion
    Pikeman bonus множитель == 2.5 vs Cavalry.
  - `internal/render/common_test.go` — после рендера common.md есть
    `### Poisoned` и `### Burning` с обоими tooltip'ами.
- `golden_test.go` остаётся структурным (length ≥ 700, наличие
  обязательных секций; список ожидаемых строк не меняется — проверки
  идут на уровне сабтестов выше).

### Не включено

- Авры зданий (`auras/structures/`) — эпик 4b.
- Loot, Trickle, Dropsite, Repair — эпик 4b.
- Healer-таблица (Heal/Range, Heal/Health, Heal/RepeatTime) — эпик 4b.
- Turrets / Projectiles (BuildingAI/DefaultArrowCount и пр.) — эпик 4b.
- Run speed (`UnitMotion/RunMultiplier`) — эпик 4b.
- Локализация имён через `.po` (gettext) — эпик 5.
- Иконки в structree — эпик 5.
- Skeleton.tmpl миграция — эпик 5/6.
- Cost batch detail — отвергнуто в `analysis.md` gap 19.
- `Stackability` в ApplyStatus — пропускаем как избыточное в MVP
  (поле фигурирует в данных, но без него unit-строка остаётся
  читаемой; добавим только если заметим расхождения).

## Архитектура

### Изменённые модули

```
internal/render/
  ├── format.go           CHANGED: + formatAttackBonuses,
  │                                + formatPreferredClasses,
  │                                + formatSplash,
  │                                + formatCaptureAttack,
  │                                + formatApplyStatuses(modeEl, byCode),
  │                                + FormatCaptureResistance,
  │                                + FormatStatusEffectResistance.
  ├── format_test.go      NEW: unit-тесты для всех новых хелперов.
  ├── units.go            CHANGED: renderUnitBlock пишет новые строки в
  │                                фиксированном порядке (см. ниже).
  ├── structree.go        CHANGED: renderBuilding добавляет строки
  │                                Resistance Capture / StatusEffect.
  ├── common.go           CHANGED: commonStatusEffects дополняется
  │                                per-status заголовками `### Poisoned`,
  │                                `### Burning` (anchor target).
  ├── common_data.go      без изменений (StatusEffect struct уже есть).
  └── report.go           CHANGED: Generator получает поле
                                   StatusEffects map[string]*StatusEffect,
                                   заполняемое в Generate
                                   через loadStatusEffects.

internal/i18n/
  ├── ru.go               CHANGED: DamageType добавляет ключи
                                   "Fire" → "огонь", "Poison" → "яд".
```

### Generator расширение

```go
type Generator struct {
    // ... существующие поля ...
    StatusEffects map[string]*StatusEffect // code → effect, lazy in Generate
}

// Generate: после LoadCiv, до renderStructree:
if g.StatusEffects == nil {
    list, err := loadStatusEffects(g.Layout.StatusEffects())
    if err != nil { return Output{}, err }
    g.StatusEffects = map[string]*StatusEffect{}
    for _, e := range list {
        g.StatusEffects[e.Code] = e
    }
}
```

`loadStatusEffects` уже существует и используется
`commonStatusEffects` — рефактор: общая карта строится один раз,
обе подсистемы её читают.

### Порядок строк в `renderUnitBlock`

```
| Стоимость              |
| Время тренировки       |
| ОЗ                     |
| Броня (H/P/C)          |
| Сопротивление захвату            (NEW, если присутствует)
| Сопротивление статус-эффектам    (NEW, если присутствует)
| Скорость               |
| Обзор                  |
| Население              |

для каждого attack-mode m в [Melee, Ranged, Capture]:
  | Атака (m)            (если у m есть Damage; для Capture — поле <Capture>)
  | Бонусы (m)           (NEW, если <Bonuses>)
  | Предпочитает (m)     (NEW, если <PreferredClasses>)
  | Брызги (m)           (NEW, если <Splash>)
  | Накладывает (m)      (NEW, по строке на статус)

| Промоушн               |
| Сбор ресурсов          |
```

Каждая «опциональная» строка пропускается, если в шаблоне нет
соответствующего под-узла, что наследует поведение существующих
строк (Атака, Население, Гарнизон).

### Порядок строк в `renderBuilding`

После строки `Броня` вставляются:

```
| Сопротивление захвату            (если присутствует)
| Сопротивление статус-эффектам    (если присутствует)
```

Прочие attack-/ApplyStatus-блоки у зданий **в 4a не реализуются** —
у зданий attack/ApplyStatus встречаются только в template_structure.xml
с `BlockChance=1, Duration=0` (no-op default), и интерпретация для
listing'а сомнительная. Если позже потребуется (gap 16 в 4b — стрелы
и turrets), вернёмся к зданиям.

### Спецификации форматтеров

#### `formatAttackBonuses(modeEl *tmpl.Element) string`

Вход: узел `Attack/Melee` (или Ranged/Capture). Возвращает строку,
пригодную для второй колонки таблицы, или `""` если узла `Bonuses`
нет / нет валидных бонусов.

Алгоритм:
1. `bonuses := modeEl.Get("Bonuses")`. Если nil — return "".
2. Для каждого `child` в `bonuses.Children` (порядок XML):
   - `classes := child.GetTokens("Classes")`. Пусто — пропуск.
   - `mul, ok := tmpl.ParseFloatTrim(child.GetText("Multiplier"))`.
     Не парсится — пропуск.
   - Сформировать `×<mul> vs <c1>+<c2>+...` (через i18n.FormatNumber
     для множителя).
3. Соединить через `, `. Пусто → return "".

Имя XML-узла (`BonusCavMelee`) не используется. Игровой движок в
`Damage.js#getAttackBonus` тоже игнорирует имя — это просто
человекочитаемый id для merge'а через `op="add"`.

#### `formatPreferredClasses(modeEl *tmpl.Element) string`

Вход: узел attack-mode. Возвращает значение второй колонки или "".

Алгоритм:
1. `tokens := modeEl.GetTokens("PreferredClasses")`. Пусто → "".
2. Возвращает `strings.Join(tokens, ", ")`.

#### `formatSplash(modeEl *tmpl.Element) string`

Вход: узел attack-mode. Возвращает строку или "".

Алгоритм:
1. `splash := modeEl.Get("Splash")`. nil → "".
2. Собрать `dmgs` из `Damage/Hack`, `Damage/Pierce`, `Damage/Crush`,
   `Damage/Fire`, `Damage/Poison` (новый набор) — i18n-форматер уже
   умеет HPC, расширяем.
3. Из `Splash/Shape` → `Circular`/`Linear` → `круг`/`линия`. Если
   отсутствует — `круг` (дефолт движка).
4. `Splash/Range` → `R=<value>`.
5. `Splash/FriendlyFire`: `false` → `не задевает союзников`,
   `true` → `задевает союзников`. Отсутствует — `задевает союзников`
   (дефолт движка `true`).
6. Suffix `, ×N vs Class` для `Splash/Bonuses` (если есть, через
   тот же `formatAttackBonuses`).
7. Итог: `<dmgs>, <shape> R=<r>, <ff-suffix>[, <bonuses>]`.

Пример: `Crush 120, круг R=1.5, не задевает союзников`.

#### `formatCaptureAttack(attackEl *tmpl.Element) string`

Вход: узел `Attack` (родитель Melee/Ranged/Capture). Возвращает
значение второй колонки строки `Атака (захват)`.

Алгоритм:
1. `cap := attackEl.Get("Capture")`. nil → "".
2. `rate, ok := tmpl.ParseFloatTrim(cap.GetText("Capture"))`. !ok → "".
3. Префикс: `захват <rate>`.
4. Скобки: `(<MaxRange>м, <RepeatTime>мс)` если оба заданы;
   деградирует — пишем то, что есть, тем же синтаксисом, что и
   у Melee/Ranged.
5. Suffix `; исключает: A, B, C` если `RestrictedClasses` непуст.
6. Suffix `; ×N vs Class` для `Capture/Bonuses` (бесплатно через
   `formatAttackBonuses` — если внезапно появятся).
7. Suffix `; накладывает: ...` для `Capture/ApplyStatus` **не**
   формируется здесь — отдельные строки `Накладывает (захват)`
   собирает `renderUnitBlock` через `formatApplyStatuses`.

Пример: `захват 2.5 (4м, 1000мс); исключает: Field, Palisade, Wall`.

#### `formatApplyStatuses(modeEl *tmpl.Element, byCode map[string]*StatusEffect) []string`

Вход: узел attack-mode + карта статус-эффектов. Возвращает срез
строк (по одной на каждый статус), пустой если нет ApplyStatus.

Алгоритм:
1. `as := modeEl.Get("ApplyStatus")`. nil → return [].
2. Для каждого `child` в `as.Children`:
   - `code := child.Name` (CamelCase: `Poisoned`, `Burning`).
   - **Edge case BlockChance:** если у child есть `BlockChance` и
     `Duration` ≤ 0 — это no-op default из template_unit/template_structure
     (мера резистентности по умолчанию). Пропускаем — это не «применение
     статуса», а фикс наследования.
   - `effect := byCode[code]`. Если nil — fallback `name = code`.
     Иначе `name = effect.StatusName` (англ. до эпика 5).
   - `damageStr := <тот же HPC-форматер расширенный Fire/Poison>`.
   - `intervalSec := <interval_ms>/1000.0`, форматируется через
     `i18n.FormatNumber` (1–2 знака).
   - `durationSec := <duration_ms>/1000.0`, аналогично.
   - Сформировать `<name>: <damageStr>/<intervalSec>с × <durationSec>с
     (см. common.md#<lower(code)>)`.
3. Возвращает срез.

Пример (мaur maiden archer):
`Poisoned: Poison 2/0.6с × 3.6с (см. common.md#poisoned)`.

Особенности:
- Если `<Damage>` у статуса отсутствует — `damageStr = ""`, итог
  `<name>: каждые <I>с × <D>с (см. ...)` (статусы вроде silenced
  без damage).
- Якорь `#<lower(code)>` соответствует тому, как GitHub генерирует id
  из заголовка `### Poisoned` (опускает в lowercase).
- Multi-status: один элемент возвращаемого среза на каждый дочерний
  элемент `ApplyStatus`. Порядок наследует XML.

#### `FormatCaptureResistance(e *tmpl.Element) string`

Вход: корневой `Element` юнита/здания. Возвращает значение второй
колонки или "".

Алгоритм:
1. `v, ok := e.GetFloat("Resistance/Entity/Capture")`. !ok → "".
2. `i18n.FormatNumber(v)`.

#### `FormatStatusEffectResistance(e *tmpl.Element) string`

Вход: корневой `Element`. Возвращает строку или "".

Алгоритм:
1. `node := e.Get("Resistance/Entity/StatusEffect")`. nil → "".
2. Для каждого `child` (например `<poisoned>0.5</poisoned>`):
   - `code := strings.ToLower(child.Name)` (lowercase для
     отличия от attack-side CamelCase).
   - `mul, ok := tmpl.ParseFloatTrim(child.Text)`. !ok → пропуск.
   - Часть строки: `<code> ×<mul>`.
3. `strings.Join(parts, ", ")`.

В R28-данных Resistance/Entity/StatusEffect отсутствует во всех
шаблонах — функция всегда возвращает "" пока кто-то не добавит
StatusEffect-резистент в шаблон. Но логика на месте, тестируется
на inline-фикстуре.

### Якоря per-status в common.md

В `commonStatusEffects` после существующей таблицы добавляется блок:

```markdown
### Poisoned

> **Применяющему:** This unit causes poison damage.
> **Пострадавшему:** This unit is poisoned.

### Burning

> **Применяющему:** ...
> **Пострадавшему:** ...
```

Заголовок берётся из `StatusEffect.StatusName` (CamelCase ≈ Code в
R28). Один заголовок на каждый файл из `data/status_effects/`.
В R28 это два файла → два под-якоря.

GitHub auto-id: `### Poisoned` → `#poisoned`. Этот id и стоит в
ссылках из ApplyStatus-строк юнитов.

### Edge case: BlockChance в base templates

`template_unit_siege.xml`, `template_structure.xml` декларируют
ApplyStatus вида:

```xml
<ApplyStatus>
  <Poisoned>
    <Duration>0.0</Duration>
    <BlockChance>1</BlockChance>
  </Poisoned>
</ApplyStatus>
```

Это «no-op default» — гарантия, что унаследовавший шаблон, добавляющий
свой ApplyStatus через `op="add"`, попадёт в правильное поле.
Семантически это **не** «эта осадка отравляет» — Duration=0 и
BlockChance=1 в сумме означают «по умолчанию ничего не происходит».

Решение в `formatApplyStatuses`: если у статус-узла одновременно
`BlockChance > 0` и (`Duration` отсутствует или `<= 0`) — пропускаем.
Также: если `Duration <= 0` без BlockChance — пропускаем (любая
ApplyStatus с нулевой длительностью бессмысленна).

Это значит, что юниты, чьи реальные ApplyStatus после наследования
дают `Duration > 0`, отрисуются корректно (Duration переопределяется
ребёнком), а base templates без override — пропустятся.

## Решения и их обоснование

### 1. Плоская 2-колоночная таблица расширяется новыми строками

Альтернатива «sub-block с вложенной таблицей» отвергнута: ломает
существующий формат, принуждает к более сложному рендеру и
переписыванию golden-тестов. Принятая схема (вариант A в
брэйнсторминге) добавляет максимум ≈3×N новых строк на attack-mode,
что комфортно вертикально читается.

### 2. Per-attack-mode prefix для всех новых строк

`Бонусы (ближ.)`, `Брызги (стрельба)`, `Накладывает (захват)` —
параллельно существующим `Атака (ближ.)` / `Атака (стрельба)`.
Однородно с текущим стилем, грep-абельно в тестах, недвусмысленно
для читателя со множественными режимами атаки у одного юнита.

### 3. CamelCase для XML-кодов статусов, lowercase для anchor

`<Poisoned>` в attack-side XML и в `code` поле JSON; `### Poisoned`
заголовок в common.md → автоматический GitHub anchor `#poisoned`.
Lowercase используется только в URL/anchor части ссылки — там, где
это правило GitHub. На уровне отображения везде оставляем CamelCase
имя статуса.

### 4. ApplyStatus номинально показывается per-status, без агрегации

Каждый дочерний элемент `<ApplyStatus>` → отдельная строка
`Накладывает (<режим>)`. Альтернатива «один row, статусы через `;`»
читается хуже и затрудняет grep в тестах.

### 5. BlockChance default — пропуск, не специальный комментарий

Альтернатива: писать строку `Накладывает (стрельба) | (заблокирован
по умолчанию)` для родительских шаблонов. Отвергнуто — это шум, не
информация: пользователь отчёта получает данные по конкретной циве,
а template_structure.xml — это движковый default, не часть листинга
цив-зданий.

### 6. Resistance/Entity/StatusEffect — lowercase коды

В `<poisoned>0.5</poisoned>` код в lower-case — это историческая
конвенция движка для resistance-side (оба варианта работают, но в
текущих fixtures и tooltips.js идёт lowercase). Отображение в
строке `Сопротивление статус-эффектам` — тоже lowercase, чтобы
различить с CamelCase именами в attack-side ApplyStatus.

### 7. Status code Fire/Poison расширяет i18n.DamageType

Вместо отдельного слоя для status-damage используем единый
`i18n.DamageType` с двумя дополнительными ключами. Аргумент:
status-damage и обычный damage одинаково парсятся (XML-элемент с
числом), и одна функция-форматер избавляет от дубликата. Если в
будущем 0ad добавит экзотический тип (`<Acid>`) — добавим ключ.

## Тестирование

Все тесты, читающие данные игры, идут через `internal/testutil` и
пропускаются на CI без `OAD_GAMEDATA_ROOT`. Pure-функции в
`format_test.go` — на inline-XML фикстурах, без gamedata.

### `internal/render/format_test.go` (NEW)

Один subtest на каждый новый форматтер, со списком кейсов:

- `formatAttackBonuses`:
  - `Bonuses` отсутствует → "".
  - Один бонус, один класс → `×2.5 vs Cavalry`.
  - Один бонус, несколько классов → `×3 vs Cavalry+Elephant`.
  - Несколько бонусов → `×2.5 vs Cavalry, ×2 vs Elephant`.
  - Бонус без `Multiplier` → пропуск.
- `formatPreferredClasses`:
  - Поле отсутствует → "".
  - Один класс → `Human`.
  - Несколько → `Human, Cavalry`.
- `formatSplash`:
  - Splash отсутствует → "".
  - Полный набор полей → `Crush 120, круг R=1.5, не задевает
    союзников`.
  - Без `FriendlyFire` → `задевает союзников` (дефолт true).
  - Linear shape → `линия`.
  - Splash + Bonuses → суффикс `, ×N vs Class`.
- `formatCaptureAttack`:
  - Capture отсутствует → "".
  - Полный набор → `захват 2.5 (4м, 1000мс); исключает: Field,
    Palisade, Wall`.
  - Без `RestrictedClasses` → без суффикса.
- `formatApplyStatuses`:
  - ApplyStatus отсутствует → пустой срез.
  - Один статус с Damage, Interval=600, Duration=3600 →
    `Poisoned: Poison 2/0.6с × 3.6с (см. common.md#poisoned)`.
  - Interval=1000 → `<dmg>/1с × <D>с` (нормализация без спец-кейса —
    цена единообразия, читаемо).
  - Несколько статусов → срез из двух элементов в порядке XML.
  - BlockChance=1, Duration=0 → пропуск (no-op default).
  - Statusless effect (no Damage) → `<name>: каждые <I>с × <D>с (...)`.
- `FormatCaptureResistance`:
  - Поле отсутствует → "".
  - `<Capture>100</Capture>` → `100`.
- `FormatStatusEffectResistance`:
  - Узел отсутствует → "".
  - `<poisoned>0.5</poisoned>` → `poisoned ×0.5`.
  - Несколько → `poisoned ×0.5, burning ×0.7`.

### `internal/render/structree_test.go` (расширения)

Структурные grep-проверки на готовом markdown цив-отчёта:

- `germ` или `spart`: для champion pikeman есть строка
  `| Бонусы (ближ.) | ×2.5 vs Cavalry |`.
- `mace` или любая siege-владеющая: для siege stonethrower есть
  `| Брызги (стрельба) | Crush 120, круг R=1.5, не задевает
  союзников |`.
- `maur`: для champion maiden archer есть
  `| Накладывает (стрельба) | Poisoned: Poison 2/0.6с × 3.6с
  (см. common.md#poisoned) |`.
- `iber`: для champion cavalry есть строка `Накладывает (стрельба) |
  Burning: Fire 2/3с × 9с (см. common.md#burning)`.
- Любая цива: для базового пехотинца есть
  `| Атака (захват) | захват 2.5 (4м, 1000мс); исключает: Field,
  Palisade, Wall |`.

### `internal/render/numeric_test.go` (расширения)

Спот-чек: для спартанского champion Pikeman парсинг даёт `Multiplier
= 2.5` против класса `Cavalry`. Проверка через прямое чтение шаблона,
не через рендер.

### `internal/render/common_test.go` (расширения)

После рендера common.md:
- Присутствует таблица «Статус-эффекты».
- Присутствуют `### Poisoned` и `### Burning` под-заголовки с обеими
  tooltip-строками.

### `golden_test.go`

Список ожидаемых секций (length ≥ 700) сохраняется. Никаких byte-diff
проверок. Untracked golden-файлы регенерируются локально и
проверяются глазами в рамках приёмки эпика.

## План коммитов (для subagent-driven dev)

Согласно памяти `feedback_workflow_subagent_driven` — пары «main + nits»:

1. **format: attack-mode helpers (Bonuses/PreferredClasses/Splash/Capture)**
   — `formatAttackBonuses`, `formatPreferredClasses`, `formatSplash`,
   `formatCaptureAttack` + `format_test.go` (subtests на каждый).
   Pure-функции, без зависимости на каталог.
   Nits: gofmt, doc-comments, edge-case кейсы.

2. **render: ApplyStatus + StatusEffects map**
   — `formatApplyStatuses`, расширение `Generator.StatusEffects`,
   рефактор `commonStatusEffects` под общую карту.
   `i18n.DamageType` дополняется ключами `Fire`, `Poison`.
   В `commonStatusEffects` добавляется генерация per-status
   заголовков `### <Code>`.
   `format_test.go` пополняется subtest'ом для `formatApplyStatuses`.
   `common_test.go` проверяет наличие новых под-якорей.
   Nits.

3. **format: Resistance Capture + StatusEffect**
   — `FormatCaptureResistance`, `FormatStatusEffectResistance`
   + subtests в `format_test.go`.
   Nits: comment accuracy, lowercase casing rationale.

4. **render: integrate новые rows в renderUnitBlock + renderBuilding**
   — расстановка строк в порядке из § «Порядок строк»; `renderBuilding`
   получает Resistance Capture/StatusEffect строки. Обновление
   `structree_test.go`, `numeric_test.go` под новые проверки.
   Nits.

5. **docs: sync analysis.md/current-state.md/CLAUDE.md под эпик 4a**
   — gaps 10/11/17 переносятся в таблицу «Закрытые gaps»; в
   `current-state.md` упомянуты новые поля; в `CLAUDE.md` обновлена
   текущая структура пакетов и описание формата `renderUnitBlock`.

После 4 — регенерация untracked goldens (`make germ` / `make all-civs`)
и code-reviewer субагент. После 5 — закрытие эпика 4a.

## Открытые вопросы

Ни одного — все решения зафиксированы в брэйнсторминге.

## Приложение: проверенные данные R28

- `template_unit_champion_infantry_pikeman.xml` — Bonuses: Cavalry ×2.5,
  PreferredClasses: Human.
- `template_unit_siege_stonethrower.xml` — Splash: Crush 120, Circular,
  R=1.5, FriendlyFire=false.
- `template_unit_infantry.xml` — Capture: rate 2.5, MaxRange=4,
  RepeatTime=1000, RestrictedClasses: Field Palisade Wall.
- `units/maur/champion_maiden_archer.xml` — ApplyStatus: Poisoned,
  Damage.Poison=2, Interval=600, Duration=3600, Stackability=Stack.
- `units/iber/champion_cavalry.xml` — ApplyStatus: Burning,
  Damage.Fire=2, Interval=3000, Duration=9000, Stackability=Stack.
- `template_unit_siege.xml`, `template_structure.xml` — ApplyStatus:
  Poisoned, Duration=0, BlockChance=1 (no-op default, пропускаем).
- `data/status_effects/poisoned.json` — code: "Poisoned",
  statusName: "Poisoned", applierTooltip / receiverTooltip.
- `data/status_effects/burning.json` — аналогично.
- `data/damage_types/` — пять типов: crush, fire, hack, pierce, poison.
