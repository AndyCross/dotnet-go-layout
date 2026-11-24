# dotnet-go-layout

One small app written twice, in .NET and in Go, in one repository, so the two
project layouts can be read side by side. The write-up is on
[stepinto.dev](https://stepinto.dev/).

The app is **Tabanco**: the chalk tab of a sherry bar in Jerez. A waiter chalks
each round on the wooden bar in front of you, in tally marks, and adds it up
when you leave. Here that is a rules library (the menu, tabs, rounds, the bill
and *la penúltima*: one more of whatever you've drunk most, because nobody
orders the last one), used by two programs: an HTTP API and a `chalk` CLI.

```
$ chalk order Mari fino fino fino fino fino px chicharrones
Mari
  Fino           ||||/              5    10,00 €
  Pedro Ximénez  |                  1     2,50 €
  Chicharrones   |                  1     4,00 €
  ----------------------------------------------
  total                                  16,50 €
```

## Layout

```
contract/            shared by both: the menu, and the HTTP contract both test suites run
dotnet/
  Tabanco.slnx
  global.json        SDK version, and the test runner
  Directory.Build.props
  Directory.Packages.props
  src/
    Tabanco.Core/    the rules: Menu, Tab, Chalk, Storage/
    Tabanco.Api/     minimal API
    Tabanco.Cli/     builds chalk
  tests/
    Directory.Build.props
    Tabanco.Core.Tests/
    Tabanco.Api.Tests/
go/
  go.mod
  cmd/
    tabanco-api/     main package
    chalk/           main package
  internal/
    tab/             the rules: menu, tabs, chalk (tests beside them)
    store/           memory and file stores
    api/             the HTTP handler
go.work              so an editor opened at the root finds the Go module
```

## Run

.NET 10 SDK:

```
cd dotnet
dotnet test
dotnet run --project src/Tabanco.Api            # http://localhost:8080
dotnet run --project src/Tabanco.Cli -- order Mari fino
```

Go 1.27:

```
cd go
go test ./...
go run ./cmd/tabanco-api                        # http://localhost:8080
go run ./cmd/chalk order Mari fino
```

Both CLIs keep tabs in `chalk.json` (`--file` to change it) in the same shape, so
either can pick up a tab the other opened.

## The API

| Request | Does |
|---|---|
| `GET /menu` | The menu, in board order. Prices are in cents |
| `GET /tabs` | Every open tab, priced |
| `POST /tabs` `{"name":"Mari"}` | Opens a tab. 409 if one is already open |
| `GET /tabs/{name}` | One tab, priced |
| `GET /tabs/{name}/chalk` | The tab as it looks on the bar, as text |
| `POST /tabs/{name}/rounds` `{"items":["fino","px"]}` | Chalks a round. An unknown item is a 400 and marks nothing |
| `POST /tabs/{name}/penultima` | One more of the usual |
| `DELETE /tabs/{name}` | Settles up: the final bill, and the tab is gone |

Errors are `application/problem+json` with `title`, `status` and `detail`.

`contract/http.json` is one evening at the bar as a list of requests and the
exact answers expected. `go/internal/api` and `dotnet/tests/Tabanco.Api.Tests`
both run it, so the two servers cannot drift apart without a test failing.

## CI

Two workflows, each filtered on its own folder and `contract/`: a change to
`go/` doesn't build the .NET solution, and a change to the contract builds both.
