# Warum ich meinen Blog selbst gebaut habe

**Kategorie:** Entwicklung

Ich baue einen Blog – nicht weil ich dringend einen brauche, sondern weil ich dabei Dinge lernen und dokumentieren will, die ich bisher in dieser Kombination noch nie umgesetzt habe. Kein Tutorial, sondern ein ehrlicher Einblick in meine Entscheidungen.

## Der ehrliche Grund

Ich entwickle Software, kenne mich mit Linux aus und habe Docker in eigenen Projekten eingesetzt – Networking, Volumes, Multi-Stage Builds, das übliche. Was mir bisher gefehlt hat, ist ein Projekt das alles zusammenbringt: ein vollständiger Stack, von der Entwicklung bis zum Deployment, auf einem echten Server, mit echten Entscheidungen.

Ein Blog ist dafür eigentlich ein blödes Beispiel – es gibt tausend fertige Lösungen. Aber genau das ist der Punkt. Ich will nicht WordPress installieren, ich will verstehen wie so ein System von Grund auf aufgebaut wird und welche Entscheidungen dabei anfallen.

Ich dokumentiere das iterativ – nicht alles auf einmal, sondern Artikel schreiben wenn etwas fertig oder gelernt ist. Wenn ich etwas verbessere oder refactore, schreibe ich darüber. Die Artikel sind in drei Kategorien aufgeteilt: Infrastruktur, DevOps und Entwicklung.

Der Blog selbst folgt demselben Prinzip. Die erste Version ist bewusst minimal: die Artikel liegen als Markdown-Dateien direkt im Frontend-Source, kein Backend, keine Datenbank, kein CMS. Das ist nicht die finale Architektur – aber es ist genug um anzufangen und echten Inhalt zu veröffentlichen bevor der Stack vollständig steht. Das Design ist ebenfalls ein erster Entwurf.

Was schrittweise kommt: das Backend mit Go und Gin übernimmt die Artikel-Verwaltung, das Design wird iterativ verbessert, und Features wie Internationalisierung oder Versionierung kommen wenn sie gebraucht werden – nicht vorher.

## Die Technologien

**Backend: Go**

Go ist für mich neu. Ich komme aus Java und JavaScript, und Go ist eine andere Denkweise: kein Klassenmodell, keine Exceptions, Fehlerbehandlung als expliziter Rückgabewert. Ich habe mich dafür entschieden weil ich eine Sprache lernen wollte die mich etwas kostet – nicht weil es der einfachste Einstieg gewesen wäre.

Als Framework nutze ich Gin, für die Datenbankanbindung GORM mit PostgreSQL.

**Frontend: React, Vite, TypeScript**

React und TypeScript habe ich bereits in eigenen Projekten eingesetzt. Aber ich merke dass ich bisher vieles intuitiv gemacht habe ohne es wirklich zu verstehen. Dieses Projekt ist der Moment wo ich das ändern will – nicht einfach etwas zum Laufen bringen, sondern verstehen warum es funktioniert.

**API-First mit OpenAPI**

Die Idee: zuerst die API-Spezifikation definieren, dann implementieren. In eigenen Projekten habe ich das bisher immer umgekehrt gemacht – API organisch wachsen lassen, Dokumentation irgendwann hinterher. Dieses Mal mache ich es von Anfang an richtig, weil ich verstehen will warum dieser Ansatz in Teams Standard ist. Ein konkreter Vorteil zeigt sich im Frontend: ich generiere den TypeScript-Client automatisch aus der Spec – Änderungen an der API werden sofort als TypeScript-Fehler sichtbar, nicht erst zur Laufzeit.

**Tests**

Playwright für End-to-End-Tests, Integrationstests für die API. Nicht um Coverage-Zahlen zu optimieren, sondern um Testautomatisierung als echten Teil des Entwicklungsprozesses zu behandeln – nicht als nachträgliches Pflichtprogramm.

## Die Infrastruktur

**Server: Arch Linux auf einem Netcup VPS**

Linux kenne ich aus meiner Informatikausbildung, Arch habe ich selbst eine Weile täglich genutzt. Netcup bietet Arch auch als vorinstalliertes Image an – ich habe mich dagegen entschieden und stattdessen das ISO über VNC gebootet und alles manuell installiert. Mehr Aufwand, aber ich weiß genau was auf dem System ist.

Für einen Produktionsserver wäre Debian oder Ubuntu LTS die konservativere Wahl. Ich habe trotzdem Arch genommen: ich kenne das System, und das Risiko trägt in diesem Fall nur mein Blog. Den Unterschied zwischen einem persönlichen Projekt und produktionskritischer Infrastruktur zu kennen ist dabei genauso wichtig wie die Technologieentscheidung selbst.

Rolling Release auf einem Server klingt riskanter als es in der Praxis ist – wenn man einen vernünftigen Ablauf hat: Arch News lesen, btrfs-Snapshot ziehen, dann updaten.

**Server Hardening**

Passwort-Authentifizierung für SSH ist deaktiviert, nur Public-Key. Die Firewall läuft mit nftables. Docker ist so konfiguriert dass es keine eigenen nftables-Regeln schreibt – was sich in der Praxis als komplizierter herausgestellt hat als gedacht, dazu gibt es einen eigenen Artikel. fail2ban überwacht SSH und sperrt IPs nach zu vielen fehlgeschlagenen Versuchen automatisch.

**Docker: zwei Compose-Stacks, ein gemeinsames Netzwerk**

Ich trenne die Infrastruktur in zwei unabhängige Compose-Stacks:

- Proxy-Stack: nginx + certbot
- Blog-Stack: blog_frontend, blog_api, blog_postgres

Beide Stacks hängen an einem gemeinsamen externen Docker-Netzwerk, über das nginx die Applikations-Container per DNS erreicht. Die Datenbank hängt nur am internen Netzwerk – von nginx aus nicht erreichbar.

Die Trennung in zwei Stacks hat einen konkreten Grund: TLS-Zertifikate und Routing haben nichts mit dem Applikationscode zu tun. Wenn ich die Applikation neu deploye, soll nginx davon nichts merken müssen – und umgekehrt.

```
                    Internet
                       │
                    [nginx]  ←── certbot (TLS)
                       │
              ┌────────┴────────┐
              ▼                 ▼
      [blog_frontend]       [blog_api]
                                │
                          blog_internal
                                │
                         [blog_postgres]
```

## Was ich weglasse – und warum

Kubernetes wäre für diesen Setup Overengineering. Ein einzelner Server, ein persönlicher Blog – da braucht es kein Cluster-Management. Ich weiß wann es sinnvoll ist: mehrere Services, Horizontal Scaling, Rolling Deployments ohne Downtime. Das ist hier nicht der Fall.

Ansible und Terraform lasse ich vorerst weg. Den Server habe ich manuell provisioniert und jeden Schritt dokumentiert – nicht weil ich die Tools nicht kenne, sondern weil Automatisierung erst dann Sinn ergibt wenn man versteht was man automatisiert. Beim nächsten Server würde ich das Ansible Playbook parallel schreiben.
