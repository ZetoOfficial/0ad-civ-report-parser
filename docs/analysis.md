# Анализ пробелов

Список того, что игра показывает в экранах **Civilization Overview**
(`page_civinfo.xml`) и **Structure Tree** (`page_structree.xml`) или что
лежит в данных 0 A.D., но в текущем отчёте отсутствует или представлено
обрезанно.

Источник анализа: чтение
`gui/reference/civinfo/`, `gui/reference/structree/`,
`gui/reference/common/{ReferencePage,TemplateLister,TemplateLoader,TemplateParser}.js`,
`gui/common/tooltips.js` плюс прямые проверки в `simulation/data/`.

## Пробелы

### 1. Парные технологии не разворачиваются  *(закрыто в эпике 3, см. внизу)*

**Что есть сейчас:** в таблице «Исследует» под зданием парный тех (файл
с префиксом `pair_*` или содержащий поля `top`/`bottom`) выводится одной
строкой с пустыми ячейками — у самого pair-файла нет ни `cost`, ни
`tooltip`, ни `modifications`.

**Чего не хватает:** разворот в две альтернативы «выбрать одно из двух»
с подгрузкой обоих под-техов и рендером их параметров. В игре
`TemplateLister.compileTemplateLists` явно вызывает `loadTechnologyPairTemplate`
и удаляет сам pair-код из списка, оставляя только две ветки.

**Источник данных в 0ad:**
- `binaries/data/mods/public/simulation/data/technologies/pair_*.json`
- `binaries/data/mods/public/simulation/data/technologies/<top>.json`,
  `<bottom>.json` (имена в полях pair-файла)
- Логика разворота: `binaries/data/mods/public/gui/reference/common/TemplateLister.js`
  (`compileTemplateLists`, ветка с `isPairTech` / `loadTechnologyPairTemplate`)

**Сложность (оценка):** низкая. Поля `Pair`/`Top`/`Bottom` в `tech.Technology`
уже есть. Нужно: в `renderResearches` детектировать pair, рекурсивно
загрузить обе ветки через `Catalog.ByName`, отрендерить две строки
с пометкой «парный — выбрать одно».

**Зависит от:** —

---

### 2. Player-bound team bonus + проверка наличия player.civbonuses

**Что есть сейчас:** командный бонус в отчёт берётся как одна строка
`civ.json.TeamBonuses[0].Description`. Авр-файл с подробным
`modifications` не подгружается.

**Чего не хватает:** игра в `BonusesSubsection.update` (`civinfo`)
объединяет **три** источника бонусов: auto-research-техи +
`player.civbonuses` (ауры, прикреплённые к Player-сущности цивы) +
`civ.json.CivBonuses`. У нас в overview третьего нет, а команд-бонус
показывается как голый текст без структурированных модификаций.

**Источник данных в 0ad:**
- Player-шаблон цивы: `binaries/data/mods/public/simulation/templates/
  special/players/<civ>.xml`. У него `<Auras datatype="tokens">`-список,
  каждый токен — путь относительно `simulation/data/auras/`.
- Авры командного бонуса: `binaries/data/mods/public/simulation/data/
  auras/teambonuses/<civ>_player_teambonus.json` (15 файлов, по одному
  на циву). Пример полей: `type`, `affects`, `affectedPlayers`,
  `modifications`, `auraName`, `auraDescription`.
- Папки `data/auras/players/` **не существует**. Авры поделены
  только на `auras/structures/`, `auras/teambonuses/`, `auras/units/{heroes,
  catafalques}/`.
- Логика разделения player-аур на civbonus/teambonus:
  `gui/reference/common/TemplateParser.js:230` (`getPlayer`):
  ```js
  if (this.getAura(auraTemplateName).affectsTeam)
      parsed.teambonuses.push(...)
  else
      parsed.civbonuses.push(...)
  ```
  Признак teambonus — поле `affectsTeam: true` в самой авре.
- Фактически в текущей версии (R28) **все 15 player-шаблонов имеют
  ровно один Aura — teambonus**. Civbonuses через player-ауры в данных
  не используются: подтверждено grep по всем `special/players/*.xml`.
  Поэтому реальный gap здесь — только подгрузить teambonus-JSON и
  показать его `auraDescription` + `modifications`, а блок
  «player.civbonuses» в текущих данных останется пустым.

**Сложность (оценка):** низкая. Один новый загрузчик
`auras/teambonuses/<civ>_player_teambonus.json` + рендер в overview.

**Зависит от:** —

---

### 3. Глобальные autoResearch-техи (`unit_advanced`, `unit_elite`) игнорируются

**Что есть сейчас:** `Catalog.AllCivBonuses(civ)` сканирует **только**
подпапку `civbonuses/`. Глобальные авто-техи лежат в корне `technologies/`
и в отчёт не попадают.

**Чего не хватает:** упомянуть, что при апгрейде ранга (Basic → Advanced →
Elite) к юнитам автоматически применяется `unit_advanced.json` и
`unit_elite.json` — они дают +25% HP, +20% loot, +10% melee, −20% ranged
spread, −30% gather speed и т. д. Без этого читатель не понимает, почему
в игре статы Elite-юнита не совпадают с базовыми из отчёта.

**Источник данных в 0ad:**
- `binaries/data/mods/public/simulation/data/technologies/unit_advanced.json`
- `binaries/data/mods/public/simulation/data/technologies/unit_elite.json`
- Признак: `"autoResearch": true` без `requirements.civ` (применяется ко всем).
- Логика игры: `gui/reference/common/TemplateLoader.js:240`
  (`findAllAutoResearchedTechs`) — рекурсивно сканирует все
  `technologies/*.json` (включая `civbonuses/`), фильтрует по
  `autoResearch: true`. То есть по тому же критерию у нас уже есть
  `civbonuses`, но в корне `technologies/` есть и общие техи —
  их мы пропускаем.

**Сложность (оценка):** низкая. Отдельный сканер по корню `technologies/`,
фильтр по `AutoResearch && RequiresCiv == ""`, рендер в виде доп.секции
«Глобальные авто-эффекты при повышении ранга».

**Зависит от:** —

---

### 4. Сложные `requirements` парсятся не до конца

**Что есть сейчас:** `requirementPhase()` в рендере читает только
`req.tech` и первый `tech` внутри `req.all`. Всё остальное → «—».

**Чего не хватает:**
- `requirements.entity` — `{"class":"Village","number":5}` (фазовые техи).
  Сейчас в колонке «Фаза» для них «—».
- `requirements.any` — конструкция «или» для tech-требований.
- `requirements.all` со смешанными ветками (tech + notciv одновременно).
- Поле `requirementsTooltip` (человекочитаемая строка вроде «Requires five
  Village Structures.») не парсится.

**Источник данных в 0ad:**
- Сами JSON в `binaries/data/mods/public/simulation/data/technologies/*.json`.
- Логика игры: `gui/common/tooltips.js` функция `getRequirementsTooltip`.

**Сложность (оценка):** низкая. Расширить `requirementPhase`/добавить
`describeRequirements` с обходом вложенных структур; добавить поле
`RequirementsTooltip` в `tech.Technology`.

**Зависит от:** —

---

### 5. `specificName` технологий и фаз не используется

**Что есть сейчас:** `Technology.SpecificName` парсится в `map[string]any`,
но рендер его не читает.

**Чего не хватает:** показывать локальное имя цивы в скобках. Например,
`phase_town_generic` имеет `specificName.han = "Xìan"`, `specificName.spart =
"Kōmopolis"`. Это важная информация для рукописных отчётов.

**Источник данных в 0ad:** поля `specificName` в JSON-файлах в
`binaries/data/mods/public/simulation/data/technologies/`.

**Сложность (оценка):** низкая. Один лук-ап по коду цивы при рендере.

**Зависит от:** —

---

### 6. Per-modification `affects` игнорируется

**Что есть сейчас:** `tech.Modification` имеет только `Value`/`Multiply`/
`Add`/`Replace`. Поле `affects` на уровне отдельной модификации не парсится.

**Чего не хватает:** в JSON встречается:
```json
{ "value": "Attack/Melee/Damage/Hack", "multiply": 1.1, "affects": "Melee" }
```
Без чтения `affects` рендер пишет «рубящий урон ближнего боя +10%», не
указывая, что это применимо только к Melee-юнитам. Для модов с разным
`affects` у разных строк это вводит в заблуждение.

**Источник данных в 0ad:** сами JSON-файлы технологий и аур.

**Сложность (оценка):** низкая. Добавить поле `Affects` в `Modification`
struct (может быть `string` или `[]string` — нужно `RawMessage` или
`any`-парсинг), пробросить в `i18n.DescribeModification`.

**Зависит от:** —

---

### 7. `supersedes` / `replaces` / `replacedBy` не используются  *(закрыто в эпике 3, см. внизу)*

**Что есть сейчас:** `Supersedes` и `ReplacedBy` парсятся, но не
рендерятся. `Replaces` (массив) — поля в struct нет.

**Чего не хватает:** показывать цепочки апгрейдов и замены. Например,
`phase_town_generic.replaces = ["phase_town"]`, `phase_town_athen.supersedes
= "phase_village"`. На основе этого можно сказать «у Афин phase_town_athen
вместо phase_town_generic».

**Источник данных в 0ad:** соответствующие поля в
`binaries/data/mods/public/simulation/data/technologies/*.json`.

**Сложность (оценка):** средняя. Нужно построить граф «supersedes →
заменяемые техи» и для каждой цивы определить, какой именно вариант
phase-теха активен.

**Зависит от:** 8

---

### 8. Цив-специфичные phase-варианты не различаются  *(закрыто в эпике 3, см. внизу)*

**Что есть сейчас:** `i18n.PhaseRequirement` сводит `phase_town_athen`,
`phase_town_pers`, `phase_town_han` и `phase_town_generic` в общий «Town».
Различие в стоимостях и эффектах теряется.

**Чего не хватает:** при рендере «фазовой» строки технологии — отдельно
указать вариант для текущей цивы (если он есть) с его конкретной
стоимостью и эффектами.

**Источник данных в 0ad:** в текущей версии (R28) собственные
phase-варианты есть только у двух цив:
- `phase_town_athen.json`, `phase_city_athen.json` (Афины);
- `phase_town_pers.json`, `phase_city_pers.json` (Персия).

Файла `phase_town_han.json` в данных нет (упоминался в плане ошибочно).
Полный список всех phase-файлов в `data/technologies/`: `phase_village.json`,
`phase_town.json`, `phase_town_generic.json`, `phase_city.json`,
`phase_city_generic.json` плюс четыре civ-варианта выше.

**Сложность (оценка):** низкая. Изменить `requirementPhase` так, чтобы
при наличии файла `phase_<phase>_<civ>.json` вернуть его поверх
generic-варианта.

**Зависит от:** 7

---

### 9. Транзитивное замыкание от StartEntities не реализовано  *(закрыто в эпике 3, см. внизу)*

**Что есть сейчас:** `civdata.Buildings` и `Units` — это `filepath.Glob`
по `templates/structures/{civ}/` и `templates/units/{civ}/`.

**Чего не хватает:** игра в `TemplateLister.compileTemplateLists` идёт от
`StartEntities` цивы → находит, что они тренируют/строят/исследуют →
рекурсивно повторяет, пока не достигнет fixed point. Это даёт «истинный»
список того, что доступно в игре, и автоматически:
- исключает мёртвые шаблоны (есть файл, но не достижим из стартовой);
- включает шаблоны, которых нет в `<civ>/` (например, техи в общем
  каталоге);
- группирует WallSet-куски как одну сущность через `template.WallSet.Templates`.

**Источник данных в 0ad:**
- `binaries/data/mods/public/gui/reference/common/TemplateLister.js`
  (`compileTemplateLists`, `deriveTemplateListsFromTemplate`).
- `binaries/data/mods/public/gui/reference/common/TemplateLoader.js`
  (`deriveProduction`, `deriveBuildQueue`, `loadEntityTemplate`).

**Сложность (оценка):** высокая. Это основной алгоритмический сдвиг.
Нужно перенести логику обхода трёх типов отношений
(`Trainer/Entities`, `Builder/Entities`, `Trainer/Technologies`) с
обработкой WallSet и пар-техов.

**Зависит от:** 1, 13

---

### 10. Бонусы атаки (`Bonuses` в `<Attack>`) не рендерятся

**Что есть сейчас:** `FormatMeleeAttack`/`FormatRangedAttack` собирают
только урон, range и repeat time.

**Чего не хватает:** блок `<Bonuses>` внутри `<Melee>`/`<Ranged>` со
строками вида `<BonusCavMelee><Classes>Cavalry</Classes><Multiplier>2.5</Multiplier></BonusCavMelee>`.
В существующих рукописных отчётах это пишется как `x2.5 vs Cavalry`.

**Источник данных в 0ad:** соответствующие поддеревья XML-шаблонов
юнитов в `binaries/data/mods/public/simulation/templates/`.

**Сложность (оценка):** низкая. Дополнительный обход `Bonuses` в
форматтере атаки.

**Зависит от:** —

---

### 11. Splash damage, capture details, status effects в атаке

**Что есть сейчас:** не рендерятся.

**Чего не хватает:** `<Splash>` (radius, shape, friendlyFire, damage),
`<Capture>` (capture rate), `<ApplyStatus>` (poison, etc.). В игре —
функции `splashDetails`, `captureDetails`, `applyStatusDetails`,
`getStatusEffectsTooltip` в `tooltips.js`.

**Источник данных в 0ad:**
- XML-шаблоны юнитов с `<Splash>`/`<Capture>`/`<ApplyStatus>` в
  `<Attack>`.
- Шаблоны статус-эффектов: `binaries/data/mods/public/simulation/data/
  status_effects/`. В R28 там только два файла: `burning.json`,
  `poisoned.json`. Каждый содержит `code`, `statusName`, `icon`,
  `applierTooltip`, `receiverTooltip`. Объёмы данных применяемого эффекта
  (damage/interval/duration) приходят с самой атаки (через `<ApplyStatus>`),
  а из status_effects/ берутся только локализованные строки и иконки.
- Логика: `gui/common/tooltips.js`.

**Сложность (оценка):** средняя для splash/capture (новые поля
форматтера атаки). Низкая для status effects, поскольку статус-файлов
всего два — простая таблица соответствия.

**Зависит от:** —

---

### 12. Авры зданий (Wonder, Temple, Fortress, Forge) не выводятся

**Что есть сейчас:** `aura.ListInDir` зовётся только для героев и
катафалка. Авры зданий из `auras/structures/` не подгружаются и не
ассоциируются с зданиями.

**Чего не хватает:** у Wonder, Temple и других зданий есть `<Auras>`-
токены, ссылающиеся на JSON в `auras/structures/`. Это и есть «бонус
от Wonder» (+50 pop), «авто-лечение в радиусе Temple», «ремонт в
Arsenal» и т. п.

**Источник данных в 0ad:**
- `binaries/data/mods/public/simulation/data/auras/structures/*.json`
- `<Auras datatype="tokens">` в шаблонах зданий.

**Сложность (оценка):** низкая. Аналогично героям: при рендере здания
прочесть токены `Auras`, загрузить каждый JSON, отрендерить таблицу.

**Зависит от:** —

---

### 13. WallSet не группируется  *(закрыто в эпике 3, см. внизу)*

**Что есть сейчас:** `wall_short`, `wall_medium`, `wall_long`, `wall_gate`,
`wall_tower`, `wallset_stone` рендерятся как отдельные здания в фазе.

**Чего не хватает:** в игре они один блок «Stone Walls» с компактным
описанием частей. Источник — поле `WallSet.Templates` в шаблоне здания
+ token-список из `civ.json.WallSets`.

**Источник данных в 0ad:**
- `civ.json.WallSets` — массив имён wallset-зданий, например
  `["structures/wallset_palisade", "structures/spart/wallset_stone"]`.
- Сам wallset (например `templates/structures/spart/wallset_stone.xml`)
  устроен так:
  ```xml
  <Entity parent="template_wallset">
    <Identity>
      <Civ>spart</Civ>
      <SpecificName>Teichos</SpecificName>
    </Identity>
    <WallSet>
      <Templates>
        <Tower>structures/spart/wall_tower</Tower>
        <Gate>structures/spart/wall_gate</Gate>
        <WallLong>structures/spart/wall_long</WallLong>
        <WallMedium>structures/spart/wall_medium</WallMedium>
        <WallShort>structures/spart/wall_short</WallShort>
      </Templates>
      <MaxTowerOverlap>0.90</MaxTowerOverlap>
      <MinTowerOverlap>0.05</MinTowerOverlap>
    </WallSet>
  </Entity>
  ```
- Логика: `gui/common/tooltips.js` функция `getWallPieceTooltip`,
  `gui/reference/common/TemplateLister.js:130` (обход
  `template.WallSet.Templates` и `{civ|native}` подстановка).

**Сложность (оценка):** средняя. Надо опознавать wallset-сущности и
скрывать их составные части как самостоятельные здания.

**Зависит от:** 9

---

### 14. Loot, Upkeep, Treasure, Resource trickle, Resource supply, Dropsite, Repair

**Что есть сейчас:** не рендерятся ни в таблицах юнитов, ни в таблицах
зданий.

**Чего не хватает (по элементам шаблона, имена подтверждены через grep
по `templates/`):**
- `<Loot>` — что выпадает при убийстве (food/wood/xp). У всех юнитов
  (есть в `template_unit.xml` базе и переопределяется детьми).
- `<Upkeep><Rates>...</Rates><Interval>...</Interval></Upkeep>` —
  стоимость содержания. Компонент: `simulation/components/Upkeep.js`.
  В R28 элемент `<Upkeep>` отсутствует во всех текущих шаблонах
  (`grep <Upkeep>` пустой). Поле теоретически поддерживается движком,
  но в данных не используется — этот пункт можно не реализовывать
  пока кто-то не добавит Upkeep в шаблон.
- `<Treasure>` — только в `template_gaia_treasure.xml` (карта/гайя),
  не в шаблонах цив. Для отчёта по циве не нужно.
- `<ResourceTrickle><Rates><food/wood/stone/metal>...</Rates><Interval>` —
  медленный приток ресурсов. Подтверждено в `template_structure_wonder.xml`
  (Wonder даёт +1/сек по всем ресурсам) и в `template_player.xml` (с
  нулями по умолчанию).
- `<ResourceSupply>` — запас в ресурсной сущности. Только в
  `template_gaia_*.xml` и `template_structure_resource_field.xml`. Из
  зданий цивы — **только Field** имеет `ResourceSupply`.
- `<ResourceDropsite>` — какие ресурсы здание/юнит принимает (Storehouse —
  wood/stone/metal, Farmstead — food, CC — все).
- `<Repairable>` — есть в `template_structure.xml` (= все здания) и в
  `template_unit_ship.xml` / `template_unit_siege.xml`. То есть все
  здания, корабли и осадные могут ремонтироваться.

**Источник данных в 0ad:** XML-шаблоны зданий и юнитов; функции
`getLootTooltip`, `getUpkeepTooltip`, `getTreasureTooltip`,
`getResourceTrickleTooltip`, `getResourceSupplyTooltip`,
`getResourceDropsiteTooltip`, `getRepairTimeTooltip` в
`gui/common/tooltips.js`.

**Сложность (оценка):** низкая по каждому полю отдельно. `Upkeep`/
`Treasure` можно опустить (нет в данных или не относится к циве).

**Зависит от:** —

---

### 15. Healer-параметры не выделяются

**Что есть сейчас:** лекари классифицируются как `IsHealer`, выводятся
в отдельной секции, но в их таблице параметров нет специфичных полей
(heal range, heal HP per tick, repeat time).

**Чего не хватает:** строка с параметрами `<Heal>` (Range, Health,
RepeatTime). В игре — `getHealerTooltip`.

**Источник данных в 0ad:**
- `<Heal>` в шаблонах юнитов-лекарей в
  `binaries/data/mods/public/simulation/templates/units/`.
- Логика: `gui/common/tooltips.js` функция `getHealerTooltip`.

**Сложность (оценка):** низкая.

**Зависит от:** —

---

### 16. Turrets и projectiles из зданий

**Что есть сейчас:** не рендерятся.

**Чего не хватает:**
- `<TurretHolder><TurretPoints><One><X/><Y/><Z/></One>...</TurretPoints>
  </TurretHolder>` — точки для гарнизона-стрелков (на стенах, башнях,
  кораблях). Имя элемента подтверждено в `template_structure_defensive_outpost.xml`.
  Каждая точка — именованный child (`One`, `Two`, ...) с координатами
  `X/Y/Z`. В игре `getTurretsTooltip` суммирует количество.
- `<BuildingAI>` — `DefaultArrowCount`, `GarrisonArrowMultiplier`,
  `MaxArrowCount`. У CC, башен, форта. Существующие рукописные отчёты
  пишут «6 стрел базово, +1 за гарнизонного юнита».

**Источник данных в 0ad:**
- `<TurretHolder>` в `template_structure_defensive_outpost.xml` и других
  оборонительных шаблонах.
- `<BuildingAI>` в шаблонах структур `templates/structures/` и
  `templates/template_structure_*.xml`.
- Логика: `gui/common/tooltips.js` функции `getTurretsTooltip` (606),
  `getProjectilesTooltip` (606).

**Сложность (оценка):** низкая. Оба поля — простое чтение элементов
через `Element.Get`. Достаточно посчитать число дочерних элементов в
`TurretPoints`.

**Зависит от:** —

---

### 17. Resistance: Capture и StatusEffect не рендерятся

**Что есть сейчас:** `FormatArmor`/`FormatArmorHPC` берут только
`Resistance/Entity/Damage/{Hack,Pierce,Crush}`.

**Чего не хватает:**
- `Resistance/Entity/Capture` — захват (важно для тяжёлых юнитов и зданий).
- `Resistance/Entity/StatusEffect` — сопротивление статус-эффектам.

**Источник данных в 0ad:** XML шаблонов; функции
`getCaptureResistanceTooltip`, `getStatusEffectsResistanceTooltip` в
`tooltips.js`.

**Сложность (оценка):** низкая.

**Зависит от:** —

---

### 18. Run speed (`RunMultiplier` / отдельный run) не выводится

**Что есть сейчас:** `FormatWalkSpeed` показывает только
`UnitMotion/WalkSpeed`.

**Чего не хватает:** `UnitMotion/RunMultiplier` или вычисленный
`RunSpeed = WalkSpeed × RunMultiplier`. В существующих рукописных
отчётах часто пишут как `Walk 9.5, Run ×1.5`.

**Источник данных в 0ad:** `UnitMotion` в шаблонах юнитов; функция
`getSpeedTooltip` в `tooltips.js`.

**Сложность (оценка):** низкая.

**Зависит от:** —

---

### 19. Cost batch detail (для тренировки пачкой)

**Что есть сейчас:** `FormatCost` выдаёт стоимость одного юнита.

**Чего не хватает:** тренировка пачкой даёт скидку (или штраф) и есть
параметры (`fullBatchSize`, `buildingsCountToTrainFullBatch`,
`remainderBatch`). В игре — функция `getEntityCostTooltip` /
`getEntityCostComponentsTooltipString` (`tooltips.js:723`). Эти параметры
**не приходят из элемента шаблона напрямую** — они вычисляются и
передаются в tooltip-функции из контекста UI (текущая выбранная пачка
в очереди здания). В рукописных отчётах батч-инфы почти нет.

**Источник данных в 0ad:** для отчёта используется только базовая
стоимость одного юнита (`Cost/Resources/*`), которую мы уже выводим.
Реализовывать batch не нужно, разве что показать
`barracks_batch_training.json` тех (в `data/technologies/`) как обычное
улучшение (через таблицу «Исследует» здания).

**Сложность (оценка):** не применимо (батч-инфо не нужно для текущего
формата отчёта).

**Зависит от:** —

---

### 20. Civ Emblem / Civ Name (localized) / Civ History — живут не там, где мы искали

**Что есть сейчас:** в верхнем info-блоке отчёта используется только
`Code` и моя собственная таблица имён EN/RU. Из `civ.json` берутся только
`Code`, `Culture`, `CivBonuses`, `TeamBonuses`, `StartEntities`,
`SkirmishReplacements`, `WallSets`.

**Чего не хватает:** игра в `CivInfoPage.js` использует:
- `this.civData[civ].Name` → локализованное имя цивы;
- `this.civData[civ].History` → длинный исторический текст;
- `this.civData[civ].Emblem` → путь к иконке герба.

При этом `loadCivData()` в JS **не** читает эти поля напрямую из
`civs/<civ>.json` — в реальных JSON-файлах (`spart.json`, `germ.json`,
`athen.json`) полей `Name`/`History`/`Emblem` **нет**. Они живут в
Player-шаблоне `templates/special/players/<civ>.xml`:
```xml
<Identity>
  <Civ>spart</Civ>
  <GenericName>Spartans</GenericName>
  <History>Sparta was a prominent city-state ...</History>
  <Icon>emblems/emblem_spartans.png</Icon>
</Identity>
```
JS-движок (через `helpers/Player.js` и `Engine.GetTemplate`) при
загрузке `civData` подмешивает эти поля из Player-шаблона в объект
цивы. Поэтому правильное место для чтения — **обязательно**
`templates/special/players/<civ>.xml`, не `civs/<civ>.json`.

**Источник данных в 0ad:**
- `binaries/data/mods/public/simulation/templates/special/players/<civ>.xml`,
  узлы `Identity/GenericName`, `Identity/History`, `Identity/Icon`.
- 15 файлов, по одному на циву (включая `gaia.xml` для нейтральной).
- Логика: `gui/reference/common/TemplateLoader.js:85`
  (`loadPlayerTemplate`), `gui/reference/civinfo/CivInfoPage.js:55`
  (`CivEmblem.sprite = this.civData[civ].Emblem`), `:56` (`Name`),
  `:57` (`History`).
- `EntityTranslateKeys = ["GenericName", "SpecificName", "Tooltip", "History"]`
  (`TemplateLoader.js:335`) — `History` помечена как переводимая через
  gettext.

**Сложность (оценка):** низкая. Resolver уже умеет читать
`special/players/<civ>.xml` (он лежит под `templates/`). Нужно загрузить
Player-шаблон в `Generator.Generate` и прочитать три поля.

**Зависит от:** —

---

### 21. CivSpecific structures и Specific Technologies не выделяются в overview

**Что есть сейчас:** в overview отчёта нет блоков «Уникальные строения»
и «Уникальные технологии». В игре эти блоки есть и формируются через
фильтр по классу `CivSpecific Structure`/`CivSpecific Technology` (см.
`StructuresSubsection.js`, `TechnologiesSubsection.js`).

**Чего не хватает:** короткий список уникальных для цивы зданий и техов
в самом начале отчёта, до пофазных блоков.

**Источник данных в 0ad:**
- Класс `CivSpecific` в `Identity/Classes` шаблонов.
- Логика: `gui/reference/civinfo/Sections/Subsections/StructuresSubsection.js`
  (`IdentifyingClassList = ["CivSpecific Structure"]`).

**Сложность (оценка):** низкая. Фильтр по классу + рендер списка.

**Зависит от:** —

---

### 22. Heroes-блок в overview отсутствует

**Что есть сейчас:** информация о героях есть в Appendix («Детальная
информация по типам юнитов» → подсекция «Герои»). В overview-секции
отчёта блока «Heroes» нет.

**Чего не хватает:** короткий список героев с одной строкой описания
(имя + класс + краткое описание ауры) перед пофазными блоками. В игре —
`HeroesSubsection.js`.

**Источник данных в 0ad:**
- Шаблоны героев: `simulation/templates/units/<civ>/hero_*.xml`.
- Их ауры: `simulation/data/auras/units/heroes/<civ>_hero_*.json`.

**Сложность (оценка):** низкая. У нас данные уже есть, нужно отдельный
рендер краткого списка.

**Зависит от:** —

## Закрытые gaps

| Gap | Заголовок | Эпик | Реализация |
|-----|-----------|------|------------|
| 1 | Парные технологии | 3 | `tech.ExpandPair`; рендер двумя строками с маркером ◐ |
| 7 | supersedes/replaces/replacedBy | 3 | `tech.Index{Chain, ResolveForCiv}`; chain-suffix в таблицах исследований |
| 8 | Civ-варианты фаз | 3 | `Index.ResolveForCiv` + `Supersedes`-fallback в `requirementPhase` |
| 9 | Транзитивное замыкание | 3 | `civdata.Reach(civ, idx, resolver, catalog)`; `Buildings()`/`Units()` удалены |
| 13 | Группировка WallSet | 3 | `civdata.IdentifyWallSets`; render `### Стены` блоком |

## Принятые решения

### 1. Разделение вывода на два файла (как в игре)

Выходов на одну циву **два**, по аналогии с двумя экранами в игре:

- `<civ>_overview.md` — Civilization Overview: код, культура, эмблема,
  история, цив-бонусы, команд-бонус, blacklist техов, краткие списки
  уникальных зданий / героев / техов.
- `<civ>_structree.md` — Structure Tree: пофазные блоки зданий с
  таблицами «Тренирует»/«Исследует», приложение «Детальная информация
  по типам юнитов», сводная таблица строимых зданий.

Структура каждого формируется через свой `text/template` skeleton:
`internal/render/skeleton_overview.tmpl`,
`internal/render/skeleton_structree.tmpl`. Это закрывает Open Question
из `CLAUDE.md` про inline vs skeleton.tmpl: переходим на skeleton.

### 2. Транзитивное замыкание от StartEntities — внедряем

Вместо текущего `filepath.Glob("structures/{civ}/*.xml")` —
обход графа от `StartEntities` цивы по отношениям
`Trainer/Entities`, `Builder/Entities`, `Researcher/Technologies`
(c учётом `{civ}|{native}` подстановки и разворота `pair_*` техов),
как в `gui/reference/common/TemplateLister.js`. Это закрывает пробел 9
и заодно убирает «мёртвые» шаблоны и упрощает группировку WallSet.

### 3. Локализация через `.po` (опционально, по флагу)

- Без `--lang` — выводим оригинальные строки (английский из шаблонов).
  Это дефолт.
- С `--lang ru` — подгружаем нужный `.po` файл из
  `binaries/data/mods/public/l10n/`, при выводе переводим строки из
  списков `EntityTranslateKeys`, `TechnologyTranslateKeys`,
  `AuraTranslateKeys` (см. `docs/sources.md`). Ненайденные строки —
  fallback на оригинал.
- В MVP поддерживаем только `ru`. Архитектура должна позволять любой
  язык, для которого в `l10n/` есть файлы (`fr`, `de`, `es` и т. д.).

### 4. История цивы — опционально, по умолчанию off

Поле `Identity/History` из `templates/special/players/<civ>.xml`
включается во вкладку Overview только если в конфиге включена опция
`IncludeHistory: true`. Дефолт — false.

### 5. Парные техи — две строки в таблице с маркером «либо/либо»

В таблице «Исследует» под зданием парный тех `pair_*` разворачивается
в две строки, оба указывают друг на друга маркером (например
`◐ Civil Service (один из двух)`). Сам `pair_*`-файл не выводится.

### 6. Глобальные autoResearch — сноска над семейством юнита

`unit_advanced.json`/`unit_elite.json` (и любые другие autoResearch-техи
в корне `technologies/` без `requirements.civ`) выводятся **один раз
сверху** для всего семейства юнита (B/A/E считаются одной семьёй),
с описанием конкретного эффекта на этого юнита и условий применения
(«применяется автоматически при ранге Advanced»). Не повторяем для
каждого ранга.

### 7. Гибридная дедупликация (внутри-отчётная + общий common.md)

Двухуровневая стратегия:

1. **Внутри-отчётная** (в `<civ>_structree.md`): для семейств с
   повторяющимися параметрами — WallSet (5 кусков), три ранга юнита
   B/A/E — выводится одна базовая «общая» секция, далее перечисляются
   только отличия конкретного куска/ранга.

2. **Внешний `common.md`** — отдельный файл рядом с цив-отчётами,
   содержащий глобальные данные, не зависящие от цивы:
   - модификаторы `unit_advanced.json` / `unit_elite.json` целиком;
   - описание damage types (`data/damage_types/`);
   - описание ресурс-типов (`data/resources/`);
   - описание status_effects (`burning`, `poisoned`);
   - дефолты из `template_unit.xml`, `template_structure.xml` (для
     полей, которые большинство наследников не переопределяют).
   В цив-отчётах при описании юнита/здания вместо повторения базовых
   значений ставится ссылка на соответствующий якорь в `common.md`.

### 8. Status effects — внутри блока атаки юнита

`<ApplyStatus>` в `<Attack>` показывается строкой в детальной таблице
атаки соответствующего юнита (например, у `poison_arrows`-юнитов —
«накладывает Poisoned: −5 HP/сек 10 сек»). Отдельной секции нет,
описание самого статус-эффекта берётся из `data/status_effects/<code>.json`.

### 9. SkirmishReplacements — не реализуем

Поле `civ.json.SkirmishReplacements` остаётся не использованным.
Поле в `civdata.Civ` остаётся (парсится), но в рендер не идёт.

### 10. Иконки — опционально, по умолчанию off

Markdown-ссылки на спрайты (`![alt](Identity/Icon path)`) включаются
во вкладку Structure Tree (для каждого здания/юнита) только если в
конфиге включена опция `IncludeIcons: true`. Дефолт — false.
Путь к иконкам — относительно корня данных игры (`art/textures/...`
или подобное); конкретный префикс — TODO выяснить при реализации
этой фичи (вне текущего скоупа аналитики).

### Сквозные следствия

- **Конфигурация.** Появляется `internal/config/config.go` со
  struct'ом всех опций и `config.json` рядом с бинарником, плюс
  CLI-флаги (`--lang`, `--include-history`, `--include-icons`,
  `--config <path>`). CLI перекрывает JSON, JSON — встроенные дефолты.
  Формат конфига: JSON (stdlib-only зависимости в проекте). См.
  `docs/superpowers/specs/2026-04-26-epic-1-foundation-design.md`.
- **Имена выходных файлов.** Меняются: вместо
  `<civ>_buildings_report.md` — пара
  `<civ>_overview.md` + `<civ>_structree.md`. Поле `OutputFile` в
  `civdata.CivCode` либо удаляется, либо становится базовым именем
  без суффикса. Существующий golden-тест на `germ` инвалидируется
  (см. `current-state.md` → известные ограничения).
