# Go aus der Perspektive eines Java- und JavaScript-Entwicklers

**Kategorie:** Entwicklung

Go ist nicht schwer – aber es denkt anders. Wer von Java oder JavaScript kommt erwartet Klassen, Exceptions, und npm. Keins davon gibt es in Go. Dieser Artikel ist kein Tutorial, sondern eine Sammlung der Dinge die mich beim Einstieg überrascht oder verwirrt haben.

## Ein Ordner ist ein Package

In Java hat jede Datei ihre eigene Klasse, und Packages sind Namespaces. In Go ist ein Ordner ein Package – alle `.go`-Dateien im selben Ordner gehören zum selben Package und teilen denselben Namespace. Es ist als wäre es eine einzige große Datei.

```
api/
├── handlers.go    // package api
├── types.go       // package api
└── middleware.go  // package api
```

`handlers.go` kann direkt auf Typen aus `types.go` zugreifen ohne Import. Das fühlt sich anfangs seltsam an, macht aber große Dateien unnötig – man teilt Code einfach in mehrere Dateien auf ohne sich um Imports zu kümmern.

## Mehrere Rückgabewerte

Go-Funktionen können mehrere Werte zurückgeben. Das wird überall genutzt – vor allem für Fehlerbehandlung:

```go
func getPost(slug string) (Post, error) {
    // ...
}

post, err := getPost("mein-post")
if err != nil {
    // Fehler behandeln
}
```

Aus JavaScript kennt man das nicht – dort gibt es nur einen Rückgabewert, Fehler kommen über Exceptions oder Promises. In Go ist Fehlerbehandlung explizit: jede Funktion die fehlschlagen kann gibt einen `error` zurück, und der Aufrufer muss damit umgehen.

Das fühlt sich am Anfang repetitiv an – `if err != nil` steht überall. Mit der Zeit merkt man dass es einen zwingt über Fehlerfälle nachzudenken statt sie zu ignorieren.

## Die Funktionssyntax

Go-Funktionen sehen auf den ersten Blick ungewohnt aus – besonders Methoden:

```go
// Funktion
func greet(name string) string {
    return "Hallo " + name
}

// Methode auf einem Typ
func (s *BlogServer) GetPosts(ctx context.Context, request GetPostsRequestObject) (GetPostsResponseObject, error) {
    // ...
}
```

Die vielen Klammern am Anfang einer Methode sind der _Receiver_ – das ist das Objekt auf dem die Methode aufgerufen wird, äquivalent zu `this` in Java oder JavaScript. `(s *BlogServer)` bedeutet: diese Methode gehört zum Typ `BlogServer`, und `s` ist die Referenz auf die Instanz.

Der Unterschied zu Java: Methoden werden nicht innerhalb einer Klasse definiert, sondern außerhalb des Typs mit einem Receiver. Das bedeutet man kann Methoden auf jeden Typ setzen – auch auf einen `int` oder `string` wenn man möchte.

## Keine Klassen, keine Vererbung

Go hat keine Klassen. Stattdessen gibt es Structs und Interfaces.

```go
// Struct – Daten
type Post struct {
    ID        int
    Title     string
    Slug      string
    Content   string
    CreatedAt time.Time
}

// Interface – Verhalten
type PostRepository interface {
    FindAll() ([]Post, error)
    FindBySlug(slug string) (Post, error)
}
```

Vererbung gibt es nicht – stattdessen Composition. Ein Struct kann ein anderes einbetten:

```go
type TimestampedModel struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Post struct {
    TimestampedModel  // eingebettet – Post hat jetzt CreatedAt und UpdatedAt
    Title string
}
```

Für jemanden der von Java kommt wo Vererbungshierarchien normal sind, ist das eine Umgewöhnung. In der Praxis ist Composition meistens die bessere Lösung – Go zwingt einen dazu.

## Interfaces sind implizit

In Java implementiert man ein Interface explizit:

```java
class BlogServer implements ServerInterface { ... }
```

In Go nicht. Ein Typ implementiert ein Interface automatisch wenn er alle Methoden hat:

```go
type ServerInterface interface {
    GetPosts(ctx context.Context, request GetPostsRequestObject) (GetPostsResponseObject, error)
}

type BlogServer struct{}

// BlogServer implementiert ServerInterface – ohne es zu deklarieren
func (s *BlogServer) GetPosts(...) (...) { ... }
```

Das ist mächtiger als es klingt: man kann Interfaces nachträglich definieren ohne den ursprünglichen Typ zu ändern. Und man kann kleine, fokussierte Interfaces schreiben statt große die alles abdecken.

## Keine Exceptions

Go hat keine Exceptions. Fehler sind normale Rückgabewerte. `panic` gibt es, aber es ist für wirklich unerwartete Zustände – nicht für normales Error-Handling.

```go
// So nicht – panic ist kein try/catch Ersatz
func getPost(slug string) Post {
    post, err := db.Find(slug)
    if err != nil {
        panic(err) // falsch
    }
    return post
}

// So
func getPost(slug string) (Post, error) {
    post, err := db.Find(slug)
    if err != nil {
        return Post{}, fmt.Errorf("post nicht gefunden: %w", err)
    }
    return post, nil
}
```

`fmt.Errorf` mit `%w` wrapped den ursprünglichen Fehler – der Aufrufer kann mit `errors.Is` oder `errors.As` prüfen was genau schiefgelaufen ist.

## Keine impliziten Conversions, kein `null`

Go hat kein `null` – stattdessen gibt es den Nullwert jedes Typs. Ein `int` ist standardmäßig `0`, ein `string` ist `""`, ein Struct hat alle Felder auf ihrem Nullwert.

Für optionale Werte gibt es Pointer:

```go
type Post struct {
    ID    int
    Title string
    Slug  *string  // optional – kann nil sein
}
```

Oder in API-Responses wo ein Feld fehlen kann nutzt man Pointer auf primitive Typen. Das ist expliziter als `null` in JavaScript, aber auch fehleranfälliger wenn man vergisst auf `nil` zu prüfen.

## Das Tooling

Go kommt mit allem was man braucht:

```bash
go build    # kompilieren
go test     # testen
go fmt      # formatieren
go vet      # statische Analyse
```

Kein Prettier, kein ESLint, kein Webpack. `go fmt` ist der einzige Formatter und alle Go-Code sieht gleich aus – keine Diskussionen über Tabs vs. Spaces.

Module funktionieren über `go.mod` – ähnlich wie `package.json`, aber ohne das Chaos von `node_modules`. Dependencies werden in einem zentralen Cache gespeichert.

## Was mich am meisten überrascht hat

Die Einfachheit. Go hat bewusst wenige Features – keine Generics bis vor kurzem, keine Overloading, keine komplexen Typsysteme. Das fühlt sich anfangs wie eine Einschränkung an. Mit der Zeit merkt man dass es Code lesbarer macht: es gibt meistens nur einen Weg etwas zu tun.

Der zweite Moment: der Compiler ist ein strenger Lehrmeister. Unbenutzte Imports sind ein Fehler. Unbenutzte Variablen sind ein Fehler. Das nervt am Anfang – und hilft auf Dauer.
