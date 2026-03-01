# OpenAPI-First: Frontend und Backend aus einer Spec generieren

**Kategorie:** Entwicklung

Der klassische Weg: Backend bauen, API organisch wachsen lassen, Frontend irgendwie synchron halten, Dokumentation irgendwann hinterher. Das Ergebnis ist meistens ein Frontend das zur Laufzeit merkt dass ein Feld umbenannt wurde.

OpenAPI-First dreht das um: zuerst die Spec, dann der Code. Frontend und Backend werden beide aus derselben Quelle generiert – Änderungen an der API werden sofort als TypeScript-Fehler sichtbar, nicht erst wenn der Nutzer auf einen Fehler stößt.

## Die Spec

Die Spec ist die Single Source of Truth. Alles andere – Go-Server-Interface, TypeScript-Client – wird daraus generiert. Meine Spec für den Blog:

```yaml
openapi: 3.0.3
info:
  title: Blog API
  version: 1.0.0

paths:
  /posts:
    get:
      summary: Alle Posts abrufen
      tags: [Posts]
      responses:
        "200":
          description: Liste aller Posts
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Post"
    post:
      summary: Post erstellen (Admin)
      tags: [Posts]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/PostInput"
      responses:
        "201":
          description: Post erstellt
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Post"
        "401":
          description: Unauthorized

  /posts/{slug}:
    get:
      summary: Einzelnen Post abrufen
      tags: [Posts]
      parameters:
        - name: slug
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Post gefunden
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Post"
        "404":
          description: Post nicht gefunden

  /auth/login:
    post:
      summary: Admin Login
      tags: [Auth]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/LoginInput"
      responses:
        "200":
          description: JWT Token
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/TokenResponse"
        "401":
          description: Falsche Credentials

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    Post:
      type: object
      properties:
        id:
          type: integer
        title:
          type: string
        slug:
          type: string
        content:
          type: string
          description: Markdown content
        created_at:
          type: string
          format: date-time

    PostInput:
      type: object
      required: [title, content]
      properties:
        title:
          type: string
        content:
          type: string

    LoginInput:
      type: object
      required: [username, password]
      properties:
        username:
          type: string
        password:
          type: string

    TokenResponse:
      type: object
      properties:
        token:
          type: string
```

Die Spec liegt im Backend-Repo unter `api/openapi.yaml`. Das Frontend referenziert sie von dort – eine Datei, zwei Generatoren.

## Backend: oapi-codegen

oapi-codegen generiert aus der Spec ein Go-Interface das implementiert werden muss. Der Vorteil gegenüber einem vollständigen Code-Generator: es wird kein fertiger Handler generiert, sondern nur das Interface und die Typen. Die Businesslogik bleibt komplett in eigener Hand.

### Installation

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

### Konfiguration

oapi-codegen wird über eine Config-Datei gesteuert. `api/oapi-codegen.yaml`:

```yaml
package: api
generate:
  - gin-server
  - types
  - spec
output: api/api.gen.go
```

`gin-server` generiert das Server-Interface und die Router-Registrierung für Gin. `types` generiert die Go-Structs für alle Schemas. `spec` bettet die OpenAPI-Spec als String ein – nützlich für automatische Dokumentation.

### Generieren

```bash
oapi-codegen --config api/oapi-codegen.yaml api/openapi.yaml
```

Das erzeugt `api/api.gen.go` mit dem Interface:

```go
type StrictServerInterface interface {
    GetPosts(ctx context.Context, request GetPostsRequestObject) (GetPostsResponseObject, error)
    PostPosts(ctx context.Context, request PostPostsRequestObject) (PostPostsResponseObject, error)
    GetPostsSlug(ctx context.Context, request GetPostsSlugRequestObject) (GetPostsSlugResponseObject, error)
    PostAuthLogin(ctx context.Context, request PostAuthLoginRequestObject) (PostAuthLoginResponseObject, error)
}
```

### Implementieren

Jetzt muss nur noch das Interface implementiert werden:

```go
type BlogServer struct {
    db *gorm.DB
}

func (s *BlogServer) GetPosts(ctx context.Context, request api.GetPostsRequestObject) (api.GetPostsResponseObject, error) {
    var posts []model.Post
    s.db.Find(&posts)

    var result []api.Post
    for _, p := range posts {
        result = append(result, api.Post{
            Id:        &p.ID,
            Title:     &p.Title,
            Slug:      &p.Slug,
            Content:   &p.Content,
            CreatedAt: &p.CreatedAt,
        })
    }

    return api.GetPosts200JSONResponse(result), nil
}
```

Der Compiler prüft dass alle Interface-Methoden implementiert sind. Wenn ein Endpoint in der Spec hinzukommt und das Interface erweitert wird, baut das Backend nicht mehr bis die Methode implementiert ist.

### In Gin registrieren

```go
func main() {
    r := gin.Default()
    server := &BlogServer{db: db}
    api.RegisterHandlersWithOptions(r, api.NewStrictHandler(server, nil), api.GinServerOptions{})
    r.Run(":8080")
}
```

## Frontend: openapi-typescript

openapi-typescript generiert TypeScript-Typen aus der Spec. Kein vollständiger Client – nur die Typen. Für die eigentlichen HTTP-Requests wird `openapi-fetch` genutzt, eine typsichere fetch-Wrapper-Library vom selben Team.

### Installation

```bash
npm install -D openapi-typescript
npm install openapi-fetch
```

### Generieren

```bash
npx openapi-typescript ../backend/api/openapi.yaml -o src/api/types.gen.ts
```

Das erzeugt `src/api/types.gen.ts` mit allen Typen aus der Spec – Schemas, Request-Bodies, Response-Typen.

### Client einrichten

`src/api/client.ts`:

```typescript
import createClient from "openapi-fetch";
import type { paths } from "./types.gen";

const client = createClient<paths>({
  baseUrl: import.meta.env.VITE_API_URL,
});

export default client;
```

### Verwenden

```typescript
import client from "@/api/client";

// Alle Posts laden
const { data, error } = await client.GET("/posts");
// data ist Post[] – vollständig getypt

// Post erstellen
const { data: newPost, error } = await client.POST("/posts", {
  body: {
    title: "Neuer Post",
    content: "# Inhalt",
  },
  headers: {
    Authorization: `Bearer ${token}`,
  },
});
```

Wenn in der Spec `title` von `string` auf `integer` geändert wird, schlägt der TypeScript-Compiler sofort an – nicht erst zur Laufzeit.

## Automatisierung

Beide Generatoren sollten nicht manuell ausgeführt werden müssen. Im Backend ein `go generate`-Kommentar in `api/generate.go`:

```go
//go:generate oapi-codegen --config oapi-codegen.yaml openapi.yaml
```

```bash
go generate ./api/...
```

Im Frontend ein npm-Script in `package.json`:

```json
{
  "scripts": {
    "generate:api": "openapi-typescript ../backend/api/openapi.yaml -o src/api/types.gen.ts"
  }
}
```

```bash
npm run generate:api
```

Die generierten Dateien (`api.gen.go`, `types.gen.ts`) werden ins Repo committed – so sieht man im Diff direkt was sich an der API geändert hat.

## Was ich gelernt habe

Der größte Vorteil zeigt sich nicht beim ersten Endpoint, sondern beim zweiten Refactoring. Wenn ein Schema umbenannt oder ein Feld geändert wird, bricht der Build an allen Stellen gleichzeitig – Backend und Frontend. Das ist keine schlechte Sache, das ist der Punkt. Der Compiler findet Inkonsistenzen die man sonst erst zur Laufzeit bemerkt.

Was mich überrascht hat: oapi-codegen generiert kein fertiges Backend, sondern nur das Interface. Das fühlt sich erst wie mehr Arbeit an, ist aber die richtige Entscheidung – die Businesslogik bleibt lesbar und testbar, ohne generiertem Code dazwischen.
