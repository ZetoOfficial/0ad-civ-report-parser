# 0ad-civ-report-parser

Детерминированный CLI на Go, который собирает русскоязычные markdown-отчёты по
цивилизациям 0 A.D. напрямую из XML/JSON шаблонов (Pyrogenesis-движок).

Замена для скилла `/civ-report`: вместо парсинга через LLM — статический
анализатор, который читает `binaries/data/mods/public/simulation/` и выдаёт
детерминированный отчёт за миллисекунды и без расхода токенов.

## Возможности

- Резолвер XML-шаблонов: множественное наследование (`parent="a|b|c"`),
  миксины, `op="mul"`/`op="add"`, `replace=""`, `datatype="tokens"` с `-`
  префиксом для удаления, подстановка `{civ}`/`{native}`.
- Группировка зданий по фазам (Village/Town/City), фиксированный порядок
  внутри фазы.
- Таблицы «Тренирует» (юниты) и «Исследует» (технологии) под каждым зданием.
- Детальное приложение по юнитам: пехота, конница, чемпионы, герои с аурами,
  осадные орудия, корабли, поддержка, катафалк.
- Цив-бонусы (auto-research), список явно недоступных через `notciv` техов.
- Сводная таблица всех зданий цивы в конце отчёта.

## Установка

```bash
go install github.com/ZetoOfficial/0ad-civ-report-parser/cmd/civreport@latest
```

или собрать локально:

```bash
go build -o civreport ./cmd/civreport
```

## Использование

```bash
civreport spart                              # → spartans_buildings_report.md
civreport спарт                              # alias-резолюция (русские/английские)
civreport --out /tmp/r.md spart              # явный путь
civreport --gamedata /path/to/0ad/binaries/data/mods/public spart
civreport --all                              # все 15 цивилизаций
civreport --check                            # smoke-тест без записи файлов
```

Поиск пути к данным игры:
1. флаг `--gamedata`,
2. переменная `OAD_GAMEDATA_ROOT`,
3. дефолт `/Users/zeto/Projects/study/0ad/binaries/data/mods/public`.

## Поддерживаемые цивилизации

`athen`, `brit`, `cart`, `gaul`, `germ`, `han`, `iber`, `kush`, `mace`, `maur`,
`pers`, `ptol`, `rome`, `sele`, `spart`. Также понимает русские и английские
алиасы (`спартанцы`, `Athenians`, `карфагеняне`, ...).

## Тесты

```bash
go test ./...
```

Тесты включают:

- **resolver_test.go** — проверка инхеританса (HP, Hack у Spartan spearman),
  `op="mul"` для cavalry walk speed, удаление токенов через `-prefix`.
- **numeric_test.go** — точечные проверки числовых полей через resolver.
- **golden_test.go** — smoke-тест отчёта по `germ`: длина ≥ 700 строк, все
  обязательные секции присутствуют.

## Архитектура

```
internal/
  paths/      — конфигурация путей до данных игры
  tmpl/       — XML-резолвер: parser, index, merge с op=, datatype=tokens
  civdata/    — civ JSON, перечисление зданий/юнитов, фазы, алиасы
  tech/       — технологии (загрузка, notciv, civbonuses)
  aura/       — авры (heroes, catafalque, structures)
  i18n/       — русские термины и переводчик modifiers → текст
  render/     — markdown-рендер отчёта
cmd/civreport/main.go — CLI
testdata/golden/      — эталонный отчёт по germ
```

## Известные ограничения

- Секция «Три ключевые уникальные черты» не генерируется (её нужно писать
  вручную или через отдельный LLM-вызов).
- Для зданий с парными технологиями (`*_pair_*.json`) специальная разметка
  «выбрать 1 из 2» не выводится — обе технологии показываются как обычные.
- Имена юнитов оставлены на латинице (`SpecificName` в шаблоне), русский
  перевод не применяется.
- Описания эффектов (`tooltip`/`auraDescription`) приходят на английском
  напрямую из шаблонов — переводятся только пути модификаций.
