# Эпик 1 — Foundation: двухфайловый вывод, config, skeleton

> Дизайн-документ. Принят в брэйнсторминге 2026-04-26.
> Часть дорожной карты из `docs/analysis.md` → «Принятые решения».

## Контекст

Утилита `civreport` сейчас генерирует один markdown-файл
`<civ>_buildings_report.md`. Целевая архитектура — два файла на циву
(`<civ>_overview.md` + `<civ>_structree.md`) плюс общий `common.md`,
рендер через `text/template` skeleton, конфиг через YAML + CLI.

Эпик 1 — **фундамент**: разрезаем рендер на два потока, вводим
инфраструктуру (config/skeleton/Makefile), но **не добавляем ни одного
нового парсера или поля**. Цель — переключить «трубу» вывода с
сохранением текущих данных. Содержательное наполнение — эпики 2–4.

## Скоуп

### Включено

- `internal/config/config.go` — struct `Config` с финальным набором
  полей (`Gamedata`, `OutDir`, `Lang`, `IncludeHistory`, `IncludeIcons`).
  YAML-loader через stdlib (`encoding/json`-аналог нужного формата —
  см. ниже про выбор формата). CLI-биндинг только для реализованных
  в эпике 1 флагов.
- `internal/render/skeleton/` — три `text/template` шаблона
  (`overview.tmpl`, `structree.tmpl`, `common.tmpl`), встроенные через
  `//go:embed`. Тонкая обёртка: info-блок + один слот `{{.Body}}`.
- Разнос `internal/render/report.go` на `overview.go`, `structree.go`,
  `common.go`. Никаких содержательных изменений в порядке секций или
  полях.
- `civdata.CivCode.OutputFile` переименовывается в `BaseName`. Имена
  файлов жёстко конструируются.
- CLI: `--out-dir <dir>`, `--config <path>` (новые); `--gamedata`,
  `--all`, `--check` (существующие); `--out` (одно-файловый) удаляется.
- `Generator.Generate` начинает возвращать структуру с двумя строками
  (`Overview`, `Structree`); рендер `common.md` отдельным методом.
- `cmd/civreport/main.go` пишет три файла в `<out-dir>/`.
- Makefile обновляется: новые имена в `clean`, новая цель
  `make golden-diff CIV=<code>`.
- `internal/render/golden_test.go` переписывается под структурный
  тест нового формата (без strict-diff).
- Удаляется `testdata/golden/germans_buildings_report.md` (legacy).
- README.md обновляется с новым CLI и форматом вывода.

### Не включено

- Любые **новые** парсеры: Player template, team-bonus aura,
  autoResearch global techs, requirements deep-parsing,
  `specificName`, `affects` per-mod, supersedes/replaces, transitive
  closure, WallSet grouping, pair-tech expansion, building auras,
  attack bonuses/splash/capture/status, loot/trickle/dropsite,
  healer/turrets/projectiles, capture-resistance, run speed, Promotion
  колонка, batch time, garrison regen, ResourceSupply MaxRecharge.
- Любые новые секции в overview или structree кроме того, что уже
  выводится текущим `report.go`.
- Обогащение `common.md` — в эпике 1 это файл с заголовком и
  TODO-комментарием; наполняется в эпике 2.
- Локализация `.po`, иконки. Поля `Lang`, `IncludeHistory`,
  `IncludeIcons` присутствуют в `Config` и YAML, но никто их не
  читает.

## Решения, принятые в брэйнсторминге

1. **Скоуп — чистый разнос.** Никакого нового содержимого, только
   разделение существующего по двум файлам.
2. **Skeleton — тонкая обёртка.** Info-блок + один `{{.Body}}`. Весь
   markdown собирается в Go в `strings.Builder` и подставляется в
   шаблон одной строкой.
3. **`CivCode.OutputFile` → `BaseName`.** Имена файлов жёстко
   конструируются: `<BaseName>_overview.md`, `<BaseName>_structree.md`,
   `common.md`. CLI: `--out-dir <dir>` для общего каталога; `--out`
   удаляется.
4. **Config struct — финальный.** Все согласованные поля сразу есть в
   struct и YAML, CLI-флаги добавляются по мере появления логики.
5. **Golden-стратегия — структурные тесты.** Без strict-diff. Утилита
   `make golden-diff CIV=germ` показывает зазор информативно, не
   fail'ит билд. Strict-diff включается в эпике 4.
6. **Миграция — пять последовательных коммитов в одном PR.** Каждый
   компилируется, тесты проходят.
7. **Эмблема цивы и «реальный эффект» в team-bonus** — не
   реализуются. Зафиксировано как known limitation.

## Архитектура

### Поток данных

```
flags + env + config.yaml
        │
        ▼
config.Load() ──→ internal/config/config.go
        │         (CLI > YAML > defaults)
        ▼
render.NewGenerator(layout, resolver)
        │
        ▼
Generator.Generate(civInfo) → render.Output{
    Overview:  string,    // готовый markdown для _overview.md
    Structree: string,    // готовый markdown для _structree.md
}
        │
        ▼
skeleton.Render(template, infoBlock, body) // text/template
        │
        ├── overview.tmpl   → итоговый _overview.md
        ├── structree.tmpl  → итоговый _structree.md
        └── common.tmpl     → итоговый common.md
                              (через Generator.RenderCommon(),
                               отдельно от Generate)
        │
        ▼
os.WriteFile × 3  в  <config.OutDir>/
```

### Целевая структура пакетов

```
cmd/civreport/main.go              CLI: парсинг флагов, режимы, запись файлов
internal/config/
  ├── config.go                    Config struct + Load(path string) (*Config, error)
  └── config_test.go               YAML парсинг, дефолты, CLI overlay
internal/paths/                    без изменений
internal/tmpl/                     без изменений
internal/civdata/
  ├── codes.go                     CivCode.OutputFile → BaseName + методы файл-имён
  ├── codes_test.go                +тест на OverviewFile()/StructreeFile()
  ├── civ.go                       без изменений
  └── enumerate.go                 без изменений
internal/tech/                     без изменений
internal/aura/                     без изменений
internal/i18n/                     без изменений
internal/render/
  ├── skeleton/
  │   ├── skeleton.go              //go:embed FS + Render(name, data) (string, error)
  │   ├── skeleton_test.go         info-блок подставляется корректно
  │   ├── overview.tmpl            тонкая обёртка
  │   ├── structree.tmpl           тонкая обёртка
  │   └── common.tmpl              тонкая обёртка
  ├── overview.go                  renderOverview() — Identity/Bonuses/Notciv
  ├── structree.go                 renderStructree() — Phases/UnitsDetail/Summary
  ├── common.go                    renderCommon() — заголовок + TODO
  ├── format.go                    без изменений (форматтеры одного поля)
  ├── report.go                    Generator + Generate() оркестратор
  ├── golden_test.go               структурный, новый формат
  └── numeric_test.go              без изменений
testdata/golden/
  ├── germans_overview.md          (существует, untracked)
  ├── germans_structree.md         (существует, untracked)
  └── germans_buildings_report.md  УДАЛЯЕТСЯ
```

### Конфиг (`internal/config/config.go`)

```go
package config

type Config struct {
    Gamedata       string `yaml:"gamedata"`
    OutDir         string `yaml:"out_dir"`
    Lang           string `yaml:"lang"`
    IncludeHistory bool   `yaml:"include_history"`
    IncludeIcons   bool   `yaml:"include_icons"`
}

func Defaults() Config {
    return Config{
        Gamedata: paths.DefaultGameDataRoot,
        OutDir:   ".",
        Lang:     "",
    }
}

// Load читает YAML и накладывает на Defaults().
// Если path пустой — возвращает Defaults() без ошибки.
func Load(path string) (*Config, error)
```

YAML-парсер: stdlib не содержит YAML. Варианты:
- (a) добавить `gopkg.in/yaml.v3` — единственная внешняя зависимость
  ради конфига.
- (b) использовать `encoding/json` и формат `config.json` вместо YAML.
- (c) минимальный самодельный парсер `key: value` (без вложенностей).

CLAUDE.md проекта говорит «зависимости: только stdlib». **Решение:
вариант (b) — JSON.** Файл — `config.json` рядом с бинарником.
Содержимое:
```json
{
  "gamedata": "/Users/zeto/Projects/study/0ad/binaries/data/mods/public",
  "out_dir": ".",
  "lang": "",
  "include_history": false,
  "include_icons": false
}
```
В CLAUDE.md и `analysis.md` упоминается `config.yaml` — это
переименовывается в `config.json` (зафиксировать в README + обновить
оба упоминания в `CLAUDE.md` и `analysis.md`).

### CLI overlay

Порядок: defaults → JSON-файл (если `--config` или существует
`./config.json`) → CLI-флаги.

В эпике 1 биндинги:
- `--gamedata <path>` → `Config.Gamedata`
- `--out-dir <dir>`   → `Config.OutDir`
- `--config <path>`   → путь к JSON-конфигу
- `--all`             → режим (не часть Config)
- `--check`           → режим (не часть Config)

В эпике 2 добавится `--include-history`, в эпике 5 — `--lang` и
`--include-icons`.

Переменная окружения `OAD_GAMEDATA_ROOT` остаётся как fallback на
случай отсутствия флага и YAML.

### Skeleton-шаблон (overview.tmpl)

```
# {{.CivName}} ({{.CivCode}}) — Civilization Overview

> Сгенерировано {{.Date}} утилитой civreport.
> Опции: include_history={{.IncludeHistory}}, include_icons={{.IncludeIcons}}, lang={{if .Lang}}{{.Lang}}{{else}}—{{end}}.
>
> Числовые значения — базовые из XML-шаблонов с применённым
> наследованием (`parent`-цепочки, миксины, `op="mul"`/`op="add"`,
> подстановки `{civ}`/`{native}`). Эффекты технологий и аур к статам
> не применены.

---

{{.Body}}
```

`structree.tmpl` симметричен (заголовок «Structure Tree» + ссылка на
`common.md`). `common.tmpl` — заголовок «Common Reference» + дисклеймер
+ `{{.Body}}` (в эпике 1 body будет «*TODO: наполняется в эпике 2*»).

### Изменения в `civdata.CivCode`

```go
// было:
type CivCode struct {
    Code       string
    NameEN     string
    NameRU     string
    OutputFile string
    Aliases    []string
}

// стало:
type CivCode struct {
    Code     string
    NameEN   string
    NameRU   string
    BaseName string  // "germans"
    Aliases  []string
}

func (c CivCode) OverviewFile() string  { return c.BaseName + "_overview.md" }
func (c CivCode) StructreeFile() string { return c.BaseName + "_structree.md" }
```

Все 15 записей в `Civilizations` обновляются. `BaseName` — это то, что
сейчас в `OutputFile` без суффикса `_buildings_report.md`.

### Generator API

```go
// internal/render/report.go
type Output struct {
    Overview  string
    Structree string
}

func (g *Generator) Generate(civInfo civdata.CivCode) (Output, error)
func (g *Generator) RenderCommon() (string, error)
```

`main.go` оборачивает `Output.Overview` и `Output.Structree` в
соответствующие skeleton-шаблоны и пишет файлы. `RenderCommon()`
вызывается один раз на запуск и тоже оборачивается через
`common.tmpl`.

### Изменения в Makefile

```makefile
clean:
	rm -f bin/civreport
	rm -f *_overview.md *_structree.md common.md
	rm -f *_buildings_report.md   # legacy

# Новая цель.
golden-diff:
	@test -n "$(CIV)" || { echo "Usage: make golden-diff CIV=<civcode>"; exit 2; }
	@./bin/civreport --out-dir /tmp/civreport-golden $(CIV)
	@base=$$(./bin/civreport --print-basename $(CIV)) && \
	 echo "=== overview diff ===" && \
	 diff -u testdata/golden/$$base'_overview.md' /tmp/civreport-golden/$$base'_overview.md' || true; \
	 echo "=== structree diff ===" && \
	 diff -u testdata/golden/$$base'_structree.md' /tmp/civreport-golden/$$base'_structree.md' || true
```

(`--print-basename` — служебная команда CLI, добавляется в эпике 1.)

### Поведение режимов

- **Одиночный** (`civreport spart`): пишет три файла в `OutDir`.
  `common.md` детерминирован, перезапись безопасна.
- **`--all`**: пишет 15×2 = 30 цив-файлов + один `common.md` в
  `OutDir`. Каждая итерация перезаписывает один и тот же
  `common.md` одним и тем же содержимым.
- **`--check`**: ничего не пишет; вызывает `Generate(civInfo)` и
  `RenderCommon()` для всех 15 цив, проверяет отсутствие ошибок и
  непустоту вывода. Адаптация: пороги длины (`>= 100` строк) проверяются
  отдельно для overview и structree (overview сейчас короткий,
  structree длинный — пороги: overview ≥ 30 строк, structree ≥ 100
  строк, заведомо выполнимо текущим контентом).

## Стратегия миграции

Эпик идёт пятью коммитами в одном PR. Каждый коммит компилируется,
`go test ./...` проходит.

### Коммит 1 — `config: add JSON loader and Config struct`

Создаёт `internal/config/config.go` с финальным набором полей и
`Load(path)`. Никто пока не вызывает. Тесты:
- `TestDefaults` — все поля имеют ожидаемые дефолты.
- `TestLoad_FileMissing` — пустой path → Defaults без ошибки.
- `TestLoad_PartialFile` — JSON с подмножеством полей; остальные
  остаются дефолтными.
- `TestLoad_BadJSON` — невалидный JSON → ошибка с понятным сообщением.

### Коммит 2 — `render/skeleton: add text/template wrappers`

Создаёт `internal/render/skeleton/{skeleton.go,*.tmpl}`. `//go:embed`,
функция `Render(name string, data any) (string, error)`. Тесты:
- `TestRenderOverview_InfoBlockSubstitution` — `{{.Date}}`,
  `{{.IncludeHistory}}`, `{{.Body}}` подставляются.
- `TestRenderStructree_*`, `TestRenderCommon_*` — аналогично.

Никто пока не вызывает.

### Коммит 3 — `render: split report.go into overview/structree/common`

Разносит существующий код на три файла. Без логических изменений в
данных. `Generator.Generate(civInfo)` начинает возвращать
`Output{Overview, Structree}`. Метод `RenderCommon()` — заглушка,
возвращает «*TODO: наполняется в эпике 2*». Тесты:
- `numeric_test.go` остаётся зелёным.
- `golden_test.go` пока на старом формате (одна строка) —
  обновляется в коммите 5.

### Коммит 4 — `civdata,cmd: rename OutputFile to BaseName, integrate config and skeleton`

Обновляет `CivCode` (поле + методы `OverviewFile()`/`StructreeFile()`),
все 15 записей в `Civilizations`. CLI:
- удаляет `--out`;
- добавляет `--out-dir`, `--config`, `--print-basename`;
- использует `Config` из коммита 1;
- оборачивает `Output.Overview`/`Output.Structree`/`RenderCommon()`
  через skeleton-шаблоны из коммита 2.

К концу коммита `civreport spart` уже пишет три файла. `--all` пишет
30 + 1. `--check` проходит.

### Коммит 5 — `makefile,test: update clean/golden targets, rewrite golden_test`

- Обновляет Makefile (`clean`, новая цель `golden-diff`).
- Удаляет `testdata/golden/germans_buildings_report.md`.
- Переписывает `golden_test.go` под новый формат: проверки наличия
  трёх файлов, info-блока, обязательных заголовков из эталона.
- Обновляет README.md.

## Тесты

| Файл | Что проверяет в эпике 1 |
|------|--------------------------|
| `internal/config/config_test.go` | Дефолты, парсинг JSON, обработка отсутствующего файла. |
| `internal/render/skeleton/skeleton_test.go` | Info-блок подставляется в трёх шаблонах. |
| `internal/civdata/codes_test.go` | Существующие случаи + `BaseName` → `OverviewFile()`/`StructreeFile()`. |
| `internal/render/numeric_test.go` | Без изменений; spot-check значений после резолва. |
| `internal/render/golden_test.go` | Переписан: для `germ` создаются три файла, оба содержат info-блок, overview содержит секции из эталона (по заголовкам), structree содержит фазовые разделы. **Strict-byte-diff не делается.** |

## Definition of Done

1. На ветке `epic-1-foundation` коммитов 5; каждый компилируется и
   `go test ./...` проходит.
2. `civreport spart` пишет `spartans_overview.md`,
   `spartans_structree.md`, `common.md` в текущую директорию.
3. `civreport --out-dir /tmp/out spart` пишет те же три файла в
   `/tmp/out/`.
4. `civreport --all` пишет 30 цив-файлов + один `common.md`.
5. `civreport --check` smoke-проверяет 15 цив без падений.
6. `civreport --config ./config.json spart` читает конфиг и
   применяет его (`OutDir` из JSON виден в info-блоке шаблона как
   подтверждение).
7. `make build`, `make test`, `make spart`, `make all-civs`,
   `make check`, `make clean` работают.
8. `make golden-diff CIV=germ` запускается и печатает diff (зазор
   ожидаемо большой).
9. `CivCode.OutputFile` удалён, `BaseName` добавлен.
10. `testdata/golden/germans_buildings_report.md` удалён; эталоны
    `germans_overview.md` и `germans_structree.md` остаются в
    `testdata/golden/`.
11. README.md и `CLAUDE.md` обновлены: новый CLI, новые имена файлов,
    `config.json` вместо упоминаемого ранее `config.yaml`.

После merge'а в `master` дорожная карта переходит к эпику 2 — наполнение
overview и `common.md`.

## Risks and mitigations

- **Содержимое overview-файла после эпика 1 будет очень бедным** —
  старый `report.go` имел секцию «общая инфо» с пятью пунктами; они
  переезжают в overview. У structree содержания больше. Это
  ожидаемо: эпик 1 не добавляет данных, только трубу.
- **Имена файлов в `Civilizations` могут потерять синхронизацию** —
  если кто-то забудет обновить запись для одной из 15 цив. Митигация:
  добавить тест, что `BaseName != ""` для всех записей.
- **Skeleton-шаблон с `{{.Body}}` может ломать markdown**, если Body
  содержит шаблонные символы (`{{`, `}}`). Митигация: в Go-коде Body
  — обычная markdown-строка, шаблон обрабатывает его как один
  токен данных, не парсит дальше. `text/template` подставляет
  через `{{.Body}}` как есть, без интерпретации.
- **YAML → JSON смена формата конфига** — расхождение с
  существующими упоминаниями в `analysis.md` и `CLAUDE.md`. Митигация:
  в коммите 5 явно правим оба документа.

## Open questions

Не осталось — все зафиксированы решениями выше или решением «не
делаем в этом эпике».
