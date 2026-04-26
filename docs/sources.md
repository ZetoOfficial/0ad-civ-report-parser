# Карта файлов 0ad

База: `/Users/zeto/Projects/study/0ad/binaries/data/mods/public/`. Все
пути ниже — абсолютные.

## Данные цивилизаций

`/Users/zeto/Projects/study/0ad/binaries/data/mods/public/simulation/data/civs/`

- Файл на цивилизацию: `<code>.json` (15 файлов: `athen.json`, `brit.json`,
  `cart.json`, `gaul.json`, `germ.json`, `han.json`, `iber.json`,
  `kush.json`, `mace.json`, `maur.json`, `pers.json`, `ptol.json`,
  `rome.json`, `sele.json`, `spart.json`).

**Поля, которые я видел в реальных файлах** (на примере `spart.json`,
`germ.json`, `athen.json`):
- `Code` — короткий код цивы (`spart`, `germ`, ...).
- `Culture` — строка ИЛИ массив строк (в `spart.json`, `athen.json`
  строка `"hele"`, в `germ.json` массив `["germ"]`). Парсится через
  `json.RawMessage`.
- `Music` — массив `{File, Type}`. Не используется.
- `CivBonuses` — массив `{Name, History, Description}`. Используется.
- `TeamBonuses` — массив с такой же структурой. В реальных JSON-файлах
  `spart.json`, `germ.json`, `athen.json` поля `TeamBonuses` **нет**;
  команд-бонус хранится отдельно в авр-файле
  `auras/teambonuses/<civ>_player_teambonus.json`. Поле в Go struct
  присутствует на случай старых версий.
- `WallSets` — массив имён wallset-зданий (например
  `["structures/wallset_palisade", "structures/<civ>/wallset_stone"]`).
- `StartEntities` — массив `{Template, Count?}`. Используется.
- `AINames` — массив имён для AI-игроков.
- `SkirmishReplacements` — `{from: to}` для подмены generic-сущностей в
  скирмише.
- `SelectableInGameSetup` — bool.
- `Region` — массив (в `germ.json` `["europe"]`). Не во всех файлах.

**Поля `Name` / `History` / `Emblem` живут НЕ ЗДЕСЬ.** В `CivInfoPage.js`
они доступны через `this.civData[civ].Name|History|Emblem`, но
наполняются из Player-шаблона `templates/special/players/<civ>.xml` через
движковый `loadCivData()`. См. раздел «Шаблоны юнитов и зданий → Player-
шаблоны цив» ниже.

## Шаблоны юнитов и зданий

База: `/Users/zeto/Projects/study/0ad/binaries/data/mods/public/simulation/templates/`

### Категории файлов

- `template_*.xml` (в корне `templates/`) — базовые шаблоны:
  `template_unit.xml`, `template_unit_infantry.xml`,
  `template_unit_infantry_melee.xml`,
  `template_unit_infantry_melee_spearman.xml`,
  `template_unit_cavalry.xml`, `template_unit_hero.xml`,
  `template_unit_hero_infantry.xml`,
  `template_unit_hero_infantry_spearman.xml`,
  `template_structure.xml`, `template_structure_military.xml`,
  `template_structure_military_barracks.xml`, и т. д.
  Я подтверждённо открывал часть этих файлов.

- `mixins/*.xml` — миксины, подключаются по имени без слеша:
  `builder.xml`, `bush.xml`, `cataphract.xml`, `chariot.xml`,
  `elephant_african.xml`, `elephant_indian.xml`, `fireship.xml`,
  `hoplite.xml`, `longsword.xml`, `merc_cav.xml`, `merc_ele.xml`,
  `merc_inf.xml`, `mercenary.xml`, `pirate.xml`, `rotary_mill.xml`,
  `shrine.xml`, `spec_champ.xml`, `trading_post.xml`. Список взят
  фактическим `ls`.

- `mixins/civ/*.xml` — цив-специфичные миксины. В мной увиденной версии
  только `athen.xml`. Используется через `parent="civ/athen"`.

- `structures/<civ>/*.xml` — здания цивы. Подтверждено для `spart`
  (33 файла, включая wallset_stone, gerousia, syssiton, persian_stoa,
  royal_stoa, theater и т. д.).

- `units/<civ>/*.xml` — юниты цивы. Подтверждено для `spart`
  (38 файлов: catafalque, cavalry_*, champion_*, hero_*, infantry_*,
  ship_*, siege_*, support_*).

- Подкаталоги корня `templates/` (полный список через `ls`):
  `birds`, `campaigns`, `gaia`, `mixins`, `obstructors`, `rubble`,
  `skirmish`, `special`, `structures`, `units`. Релевантность для отчёта
  по циве:
  - `structures/<civ>/`, `units/<civ>/` — основной источник.
  - `mixins/`, `mixins/civ/` — миксины для parent-цепочек.
  - `template_*.xml` (в корне) — базовые шаблоны.
  - `special/players/<civ>.xml` — Player-шаблон цивы (см. ниже).
  - `special/formations/`, `special/filter/`, `special/rallypoints/` —
    вспомогательные сущности (формации, фильтры, точки сбора), на
    отчёт не влияют.
  - `gaia/`, `birds/`, `obstructors/`, `rubble/` — карта/природа,
    к циве не относятся.
  - `skirmish/` — generic-замены, применяются через
    `civ.json.SkirmishReplacements`.
  - `campaigns/` — кампании, не нужны для скирмиш-отчёта.

### Player-шаблоны цив

`templates/special/players/<civ>.xml` (15 файлов: athen, brit, cart, gaul,
germ, han, iber, kush, mace, maur, pers, ptol, rome, sele, spart; плюс
`gaia.xml`).

Структура (на примере `spart.xml`):
```xml
<Entity parent="template_player">
  <Auras datatype="tokens">teambonuses/spart_player_teambonus</Auras>
  <Identity>
    <Civ>spart</Civ>
    <GenericName>Spartans</GenericName>
    <History>Sparta was a prominent city-state ...</History>
    <Icon>emblems/emblem_spartans.png</Icon>
  </Identity>
  <Player>
    <Formations datatype="tokens">special/formations/phalanx</Formations>
  </Player>
</Entity>
```

Откуда берутся:
- **Name цивы** = `Identity/GenericName`.
- **История цивы** = `Identity/History` (gettext-переводимое).
- **Emblem (герб)** = `Identity/Icon` (например `emblems/emblem_spartans.png`).
- **Командный бонус** (token из `Auras`) = ссылка на JSON в
  `auras/teambonuses/<civ>_player_teambonus.json`. Авра имеет поле
  `affectsTeam: true`, по которому `TemplateParser.getPlayer()`
  относит её к teambonuses.
- **Civ-бонусы как авры** на player-сущности — теоретически возможны
  (отделены от teambonus по отсутствию `affectsTeam: true`), но во всех
  15 текущих player-шаблонах R28 ровно один Aura — teambonus.
  Civbonuses в текущей версии живут только в
  `data/technologies/civbonuses/`.

Базовый шаблон `templates/template_player.xml` содержит общие
ограничения (`<EntityLimits>`: лимиты Wonder=1, Hero=1, и т. д.),
формации, ресурс-trickle (по нулям) и `<TechnologyManager/>`.

### Наследование (parent)

Корень файла: `<Entity parent="...">`. Атрибут `parent` имеет вид:

- `template_<name>` — резолвится в `templates/template_<name>.xml`;
- `<mixin_name>` (без слеша) — резолвится в `templates/mixins/<mixin_name>.xml`;
- `<dir>/<name>` (со слешем) — резолвится через suffix-lookup среди
  всех найденных XML, либо как путь от `templates/`;
- `parent="a|b|c"` — множественное наследование. Применяются слева
  направо: сначала `a`, затем `b`, затем `c`, затем сам файл поверх.

### Особые семантики на уровне элементов

- `op="mul"` или `op="add"` на числовом листе — арифметика с родительским
  значением (`<WalkSpeed op="mul">2</WalkSpeed>`).
- `replace=""` на контейнере — отбросить унаследованных детей и взять
  только детей текущего файла (`<Footprint replace="">`).
- `datatype="tokens"` — содержимое — whitespace-разделённый список
  токенов. При мерже: токены без префикса добавляются (если уникальны);
  токены с префиксом `-` удаляются из унаследованного списка.
- `{civ}` и `{native}` в значениях токенов — заменяются на код цивы
  при использовании (а не при резолве).

### Где какие числовые поля живут

**Юниты (после резолва родительской цепочки):**
- `Health/Max` — ОЗ.
- `Resistance/Entity/Damage/{Hack,Pierce,Crush}` — броня.
- `Attack/Melee/Damage/{Hack,Pierce,Crush}` — урон ближнего боя.
- `Attack/Melee/{MaxRange,PrepareTime,RepeatTime}`.
- `Attack/Melee/Bonuses/<BonusName>/{Classes,Multiplier}` — бонус против
  класса.
- `Attack/Ranged/...` — те же поля для стрельбы.
- `Attack/Ranged/Projectile/{Speed,Spread,Gravity,FriendlyFire,LaunchPoint}` —
  баллистика.
- `Cost/Population`, `Cost/BuildTime`, `Cost/Resources/{food,wood,stone,metal}`.
- `UnitMotion/{WalkSpeed,RunMultiplier,Acceleration,TurnRate}`.
- `Vision/Range`.
- `Identity/{Civ,Lang,GenericName,SpecificName,Icon,Classes,VisibleClasses,
  Phenotype,Rank,RequiredTechnology,Tooltip,Undeletable}`.
- `Promotion/{Entity,RequiredXp}`.
- `Loot/{food,wood,stone,metal,xp}`.
- `Auras` (token-список, ссылается на JSON в `auras/`).
- `ResourceGatherer/{MaxDistance,BaseSpeed,Rates/...,Capacities/...}`.

**Здания:**
- `Health/Max`, `Resistance/Entity/...` — то же.
- `Cost/...`, `Cost/PopulationBonus`.
- `Footprint/{Circle/radius, Square/width/depth, Height}`.
- `TerritoryInfluence/{Root,Radius}`.
- `GarrisonHolder/{Max,List,LoadingRange,Pickup}`.
- `BuildingAI/{DefaultArrowCount,GarrisonArrowMultiplier,MaxArrowCount}`.
- `Trainer/{Entities datatype=tokens, Technologies datatype=tokens}`.
- `Builder/Entities datatype=tokens` (на юнитах).
- `ProductionQueue/{Entities,Technologies,TimeMultiplier}` — на части
  шаблонов вместо/вместе с `Trainer`.
- `Vision/Range`.
- `Identity/{Civ,GenericName,SpecificName,Icon,Classes,VisibleClasses,
  RequiredTechnology}`.
- `Auras` (token-список).
- `WallSet/Templates/{Tower,Gate,WallLong,WallMedium,WallShort}` плюс
  `WallSet/{MaxTowerOverlap,MinTowerOverlap}` — для wallset-сущностей
  (например `structures/spart/wallset_stone.xml`).
- `Upgrade/<UpgradeName>/{Entity,Tooltip,Cost/{wood,stone,...},
  Time?,Icon?,RequiredTechnology?}` — возможные апгрейды. Подтверждено
  на `template_structure_defensive_tower_sentry.xml`:
  ```xml
  <Upgrade>
    <Tower>
      <Entity>structures/{civ}/defense_tower</Entity>
      <Tooltip>Reinforce with stone and upgrade to a defense tower.</Tooltip>
      <Cost><wood>50</wood><stone>100</stone></Cost>
    </Tower>
  </Upgrade>
  ```
  Плюс встречается в `template_structure_defensive_wall_long.xml`
  (стена → ворота), `template_unit_ship_fire.xml` (Ignition).
- `TurretHolder/TurretPoints/<Name>/{X,Y,Z}` — точки гарнизона-турелей
  (на `template_structure_defensive_outpost.xml` и других башнях/стенах).
- `BuildingAI/{DefaultArrowCount,GarrisonArrowMultiplier,MaxArrowCount}`
  — стрелы от здания.
- `Repairable` — есть в `template_structure.xml` (все здания),
  `template_structure_defensive_wall.xml`, `template_unit_ship.xml`,
  `template_unit_siege.xml`.
- `ResourceSupply` — только в gaia (`template_gaia_*`) и
  `template_structure_resource_field.xml`. Из зданий цивы только Field.
- `ResourceTrickle/Rates/{food,wood,stone,metal}` + `ResourceTrickle/Interval`
  — приток ресурсов. В `template_structure_wonder.xml` (Wonder)
  и `template_player.xml` (по нулям).
- `Upkeep/Rates/{food,wood,stone,metal}` + `Upkeep/Interval` — стоимость
  содержания. Компонент `simulation/components/Upkeep.js`. В R28
  элемент в данных не используется ни одним шаблоном.
- `Trader/GainMultiplier` — для торговцев
  (`template_unit_support_trader.xml`, `template_unit_ship_merchant.xml`).
- `Heal/{Range,Health,RepeatTime}` — для лекарей.

## Технологии

База: `/Users/zeto/Projects/study/0ad/binaries/data/mods/public/simulation/data/technologies/`

- Корневая папка содержит ~150+ JSON-файлов. Категории, которые я видел:
  - **Фазы:** `phase_village.json`, `phase_town.json`, `phase_town_generic.json`,
    `phase_town_athen.json`, `phase_town_pers.json`, `phase_city.json`,
    `phase_city_generic.json`, `phase_city_athen.json`, `phase_city_pers.json`.
  - **Авто-применяемые ранги:** `unit_advanced.json`, `unit_elite.json`.
  - **Парные:** `pair_*.json` (`pair_unlock_civil_service_han`,
    `pair_unlock_champions_sele`, `pair_gather_food_maur`, `pair_unlock_cult_ptol`,
    `pair_unlock_civil_engineering_han`).
  - **Цив-специфичные:** `agoge`, `immortals`, `silvershields`, `juggernauts`,
    `roman_reforms`, `iphicratean_reforms`, `hellenistic_metropolis`,
    `parade_of_daphne`, `pharaonic_cult`, `serapis_cult`, `silk_road`,
    `helot_economy`, `krypteia`, `ostracism`, `tyrtean_paeans`, и т. д.
  - **Универсальные:** `soldier_attack_melee_01..03`, `soldier_resistance_*_01..03`,
    `soldier_attack_ranged_01..03`, `cavalry_*`, `tower_*`, `siege_*`,
    `warship_*`, `gather_*`, `trade_*`, `pop_house_*`, и т. д.
  - **Анлоки:** `unlock_champion_cavalry`, `unlock_champion_chariots`,
    `unlock_champion_infantry`, `unlock_civilians_house_generic`,
    `unlock_civilians_house_kush`, `unlock_females_house`, `unlock_neodamodes`,
    `unlock_shared_dropsites`, `unlock_shared_los`, `unlock_spies`,
    `upgrade_rank_advanced_mercenary`.

`/Users/zeto/Projects/study/0ad/binaries/data/mods/public/simulation/data/technologies/civbonuses/`

- Цив-бонусы, отдельная подпапка. Не каждый файл назван `<civ>_*.json`:
  встречаются `delian_league.json` (athen), `greek_structures.json`
  (общий для греческих), `celt_structures.json` (общий для кельтов),
  плюс именованные `cart_walls`, `cart_stone_01..03`, `germ_meat`,
  `germ_women`, `gaul_cavalry`, `mace_storehouse`, `maur_elephants`,
  `maur_population`, `maur_walls`, `pers_population`, `ptol_structures`,
  `rome_siege`, `sele_farms`, `han_population`, `han_walls`, `brit_woad_warriors`.

### Структура одного JSON

Поля, которые я видел в реальных файлах:
- `genericName` — название.
- `specificName` — `{civ_code: localized_name}` (например в
  `phase_town_generic`).
- `description` — описание.
- `tooltip` — короткое объяснение эффекта.
- `cost` — `{food?, wood?, stone?, metal?}`.
- `researchTime` — секунды.
- `requirements` — рекурсивная структура:
  - `{tech: "phase_town"}` — простое требование на тех.
  - `{civ: "spart"}` — только для конкретной цивы.
  - `{notciv: "cart"}` — недоступно для цивы; может быть строкой или
    массивом.
  - `{entity: {class: "Village", number: 5}}` — N зданий определённого
    класса.
  - `{all: [...]}` / `{any: [...]}` — конъюнкция/дизъюнкция вложенных.
- `requirementsTooltip` — человекочитаемая строка для отображения
  требований (например «Requires five Village Structures.»).
- `modifications` — массив `{value, multiply?, add?, replace?, affects?}`:
  - `value` — путь типа `Health/Max`, `Attack/Melee/Damage/Hack`,
    `Cost/Resources/wood`, `ResourceGatherer/Rates/food.meat`.
  - `multiply` или `add` — операция (могут оба не присутствовать у
    `replace`-модификаций).
  - `affects` — может быть на уровне отдельной модификации или на
    уровне всего теха (поле `affects` сверху).
- `affects` — массив классов целевых сущностей (`["Soldier"]`,
  `["Advanced Unit", "Elite Unit"]`).
- `autoResearch` — bool. Если `true`, тех применяется без клика.
- `supersedes` — имя теха, который этот заменяет.
- `replaces` — массив имён техов, которые заменяет (видел в
  `phase_town_generic.json`: `["phase_town"]`).
- `top`, `bottom`, `pair` — указатели для парных техов.
- `replacedBy` — поле описано в Go struct `Technology`, но в текущей
  версии данных (R28) **не встречается** ни в одном `.json` файле в
  `simulation/data/`. Поле остаётся в struct «на будущее»; парсится в
  ноль для всех существующих техов.
- `icon` — путь к иконке.
- `soundComplete` — путь к звуку.

## Локализация имён

- **Юниты и здания:** локализация через `Identity/SpecificName` в самом
  XML-шаблоне (язык цивы, например греческий, латинский) и
  `Identity/GenericName` (английский). Имя задаётся в шаблоне, не в
  отдельном словаре.

- **Технологии:** через поле `specificName: {civ_code: name}` внутри
  каждого JSON-теха (например `phase_town_generic.json` имеет
  `specificName.spart = "Kōmopolis"`).

- **Поля, которые игра gettext-переводит** при загрузке:
  - На сущностях (`EntityTranslateKeys` в `TemplateLoader.js:335`):
    `GenericName`, `SpecificName`, `Tooltip`, `History`.
  - На технологиях (`TechnologyTranslateKeys`, `:336`): `genericName`,
    `tooltip`, `description` (плюс `specificName[civ]` для каждого
    языка цивы — отдельная ветка в `loadTechnologyTemplate`).
  - На аурах (`AuraTranslateKeys`, `:334`): `auraName`, `auraDescription`.

- **Перевод интерфейса игры:** `binaries/data/mods/public/l10n/*.po`
  (gettext PO-файлы). Для текущего отчёта не подключаем — мы пишем
  оригинальные английские имена `GenericName`/`SpecificName` плюс
  русские термины через собственный словарь `i18n.pathTranslations`.

- **Имена цивилизаций для UI** — `Identity/GenericName` в Player-шаблоне
  `templates/special/players/<civ>.xml`. См. раздел «Player-шаблоны цив»
  выше. В `civs/<civ>.json` поля `Name` нет (вопреки тому, что код GUI
  обращается к `civData[civ].Name` — это поле подставляется движком при
  `loadCivData`).

## Алгоритмы в коде

Логика, которую невозможно восстановить только из данных — она реализована
в JS (GUI) и должна быть реплицирована в Go.

- `binaries/data/mods/public/gui/reference/common/TemplateLister.js` —
  построение списка достижимых юнитов/зданий/техов транзитивным замыканием
  от `StartEntities` цивы (с разворотом парных техов и обработкой
  WallSet).

- `binaries/data/mods/public/gui/reference/common/TemplateLoader.js` —
  загрузка шаблонов и парсинг данных. Конкретные методы:
  - `loadEntityTemplate(name, civCode)` (`:52`) — делегирует
    `Engine.GetTemplate()` (C++-сторона); итоговая parent-цепочка
    разворачивается там. Затем JS делает `clone()` и
    `translateObjectKeys()` (`EntityTranslateKeys`).
  - `loadPlayerTemplate(civCode)` (`:85`) — путь
    `special/players/<civCode>` через `buildPlayerTemplateName()`.
  - `loadAuraTemplate(name)` (`:30`), `AuraPath = "simulation/data/auras/"`.
  - `loadTechnologyTemplate(name)` (`:106`),
    `TechnologyPath = "simulation/data/technologies/"`.
  - `loadTechnologyPairTemplate(name, civ)` (`:131`) — возвращает
    `{techs: [top, bottom], reqs: ...}`.
  - `findAllAutoResearchedTechs()` (`:240`) — рекурсивный обход
    `technologies/*.json` с фильтром `autoResearch: true`.
  - `deriveProduction(template, civ)` (`:140`) — вычисление того, что
    шаблон может тренировать/исследовать (с подстановкой `{civ|native}`
    и разворотом pair-техов).
  - `deriveBuildQueue(template, civ)` (`:187`) — что юнит может строить.
  - `getVariantBaseAndType(name, civ)` (`:268`) — классификация
    шаблона как `base`/`promotion`/`upgrade`/`unlockedByTechnology`/
    `trainable`/`unknown`.
  - `isPairTech(code)` (`:308`) — детекция парного теха по полю
    `template.top`.
  - `isPhaseTech(code)` (`:313`) — детекция фазового теха по префиксу
    `phase` в basename.

- `binaries/data/mods/public/gui/reference/common/TemplateParser.js` —
  применение модификаций технологий и аур к шаблонам, derive списка
  активных фаз для текущей цивы (`derivePhaseList`,
  `deriveModifications`). Это «применить эффекты технологий к статам»,
  что у нас отсутствует (мы выводим только базовые значения).

- `binaries/data/mods/public/gui/reference/common/ReferencePage.js` —
  определение списка `StatsFunctions` (16 tooltip-функций), которые
  игра вызывает для каждой сущности. Это «оглавление» того, что должно
  попасть в детальную таблицу юнита/здания.

- `binaries/data/mods/public/gui/common/tooltips.js` — реализация всех
  `getXxxTooltip` функций. Каждая отвечает за одну колонку или строку
  в табличке игры. ~50 функций, ~1262 строки.

- `binaries/data/mods/public/gui/reference/civinfo/Sections/Subsections/BonusesSubsection.js` —
  логика объединения трёх источников цив-бонусов (auto-research техи +
  Player-bound auras + `civ.json.CivBonuses`).

- `binaries/data/mods/public/gui/reference/civinfo/Sections/Subsections/HeroesSubsection.js`,
  `StructuresSubsection.js`, `TechnologiesSubsection.js` —
  фильтры по классам `Hero` и `CivSpecific Structure`/`CivSpecific
  Technology` для overview-блоков.

- `binaries/data/mods/public/gui/reference/structree/StructreePage.js` +
  `Sections/Tree/TreeSection.js` — компоновка дерева структур по фазам.

- **C++-движок** (`source/simulation2/components/CCmp*.cpp`) — реальная
  игровая логика: применение модификаторов, накопление статов,
  pathfinding. `Engine.GetTemplate()` (вызываемый из JS-`loadEntityTemplate`)
  тоже C++; именно там разворачивается полная parent-цепочка XML.

  Для нашего отчёта C++-сторона не нужна: parent/op/tokens/replace
  семантика уже реплицирована в `internal/tmpl/`, а статы выводятся
  базовые без применения техов. Залезать в `source/` может потребоваться
  только для верификации поведения `op="mul"`/`op="add"` или
  `replace=""` в спорных случаях (например, как именно мержатся
  пересекающиеся `datatype="tokens"` с одинаковыми токенами в разных
  ветках наследования).
