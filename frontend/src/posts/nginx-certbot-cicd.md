# TLS mit nginx und certbot – und wie man das Henne-Ei-Problem löst

**Kategorie:** DevOps

nginx braucht ein Zertifikat zum Starten. certbot braucht einen laufenden nginx um das Zertifikat auszustellen. Beides gleichzeitig geht nicht.

Ich kannte das Problem bereits – ich hatte nginx und certbot schon mal ohne Docker aufgesetzt und dabei dasselbe Problem gelöst. Aber dockerized ist es eine andere Sache. Die Lösung ist dieselbe, der Weg dorthin nicht.

## Die Struktur

Ich trenne Infrastruktur in zwei unabhängige Docker Compose Stacks: einen Proxy-Stack unter `/opt/proxy` und einen Blog-Stack unter `/opt/blog`. Der Proxy-Stack kennt die Applikation nicht – er routet nur Traffic und kümmert sich um TLS. Wenn ich die Applikation neu deploye, merkt nginx davon nichts.

```
/opt/proxy/
├── docker-compose.yml
├── bootstrap.sh
├── .env
├── .env.example
├── nginx/
│   ├── init.conf
│   └── conf.d/
│       └── blog.conf
└── certbot/
    └── www/
```

Die Verzeichnisse müssen vor dem ersten Start existieren:

```bash
mkdir -p nginx/conf.d certbot/www
```

Die E-Mail-Adresse für certbot landet in einer `.env` Datei – nicht hardcodiert im Skript:

```bash
# .env
CERTBOT_EMAIL=deine@email.de
```

Beide Stacks kommunizieren über ein gemeinsames externes Docker-Netzwerk:

```bash
docker network create proxy_network
```

## docker-compose.yml

```yaml
services:
  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - ./certbot/www:/var/www/certbot:ro
      - certbot_certs:/etc/letsencrypt:ro

  certbot:
    image: certbot/certbot
    volumes:
      - ./certbot/www:/var/www/certbot
      - certbot_certs:/etc/letsencrypt

volumes:
  certbot_certs:

networks:
  default:
    name: proxy_network
    external: true
```

certbot läuft hier nicht dauerhaft – dazu gleich mehr.

## Das Henne-Ei-Problem

nginx liest beim Start alle Configs in `conf.d/` ein. Wenn dort ein `ssl_certificate`-Pfad steht der noch nicht existiert, verweigert nginx den Start. Das Zertifikat kann aber erst ausgestellt werden wenn nginx läuft und HTTP-Requests beantworten kann.

Der Ausweg ist ein gestufter Bootstrap. In Stage 1 startet nginx mit einer minimalen Config ohne SSL. In Stage 2 holt certbot das Zertifikat. In Stage 3 lädt nginx die echte Config.

### nginx/init.conf

Temporäre Config für Stage 1 – matcht alle Domains und liefert nur den Challenge-Pfad aus:

```nginx
server {
    listen 80;
    server_name _;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 200 'ok';
    }
}
```

### nginx/conf.d/blog.conf

Die echte Config nach dem Bootstrap:

```nginx
server {
    listen 80;
    server_name blog.example.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    http2 on;
    server_name blog.example.com;

    ssl_certificate /etc/letsencrypt/live/blog.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/blog.example.com/privkey.pem;

    # Versionsnummer nicht in Fehlermeldungen und Response-Headern preisgeben
    server_tokens off;

    # Browser nur noch über HTTPS erlauben, für 6 Monate merken
    add_header Strict-Transport-Security "max-age=15768000; includeSubDomains" always;

    # Seite darf nicht in einem iframe eingebettet werden (Clickjacking)
    add_header X-Frame-Options "SAMEORIGIN" always;

    # Browser soll den MIME-Type nicht selbst erraten, sondern dem Server vertrauen
    add_header X-Content-Type-Options "nosniff" always;

    # Beim Navigieren auf externe Seiten keinen vollen Referrer mitschicken
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    location / {
        proxy_pass http://blog_frontend:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Der Challenge-Pfad bleibt auch in der Produktions-Config drin – certbot braucht ihn für Renewals.

### bootstrap.sh

Das Skript liest die Domains direkt aus den vorhandenen nginx-Configs aus. `conf.d/` ist die Single Source of Truth – keine separate Liste die man vergessen kann zu aktualisieren.

Beim ersten Versuch hatte ich ein Problem: `blog.conf` lag bereits in `conf.d/` als Stage 1 lief. nginx las beide Configs ein, fand den `ssl_certificate`-Pfad der noch nicht existierte, und crashte sofort. certbot konnte die Challenge nicht abrufen weil nginx nicht erreichbar war – das Henne-Ei-Problem in einer anderen Form.

Die Lösung: die echten Configs während Stage 1 temporär ausblenden und erst nach dem Zertifikats-Holen wiederherstellen.

```bash
#!/bin/bash
set -e

source .env
COMPOSE="docker compose"

echo "==> Stage 1: nginx mit init.conf starten"
mkdir -p nginx/conf.d.bak
mv nginx/conf.d/*.conf nginx/conf.d.bak/ 2>/dev/null || true
cp nginx/init.conf nginx/conf.d/init.conf
$COMPOSE down 2>/dev/null || true
$COMPOSE up -d nginx
sleep 2

echo "==> Stage 2: Zertifikate holen"
domains=$(grep -rh "server_name" nginx/conf.d.bak/*.conf \
    | grep -v "_;" \
    | awk '{print $2}' \
    | tr -d ';' \
    | sort -u)

for domain in $domains; do
    echo "  -> Zertifikat für $domain"
    $COMPOSE run --rm certbot certonly \
        --webroot \
        --webroot-path=/var/www/certbot \
        --email "$CERTBOT_EMAIL" \
        --agree-tos \
        --no-eff-email \
        --keep-until-expiring \
        -d "$domain"
done

echo "==> Stage 3: Echte Configs wiederherstellen, nginx neu laden"
mv nginx/conf.d.bak/*.conf nginx/conf.d/
rm -rf nginx/conf.d.bak
rm -f nginx/conf.d/init.conf
$COMPOSE exec nginx nginx -s reload

echo "==> Fertig"
```

Wenn später eine neue Domain dazukommt: Config anlegen, `bootstrap.sh` nochmal ausführen. `--keep-until-expiring` sorgt dafür dass certbot ein bereits gültiges Zertifikat nicht neu anfordert.

## Renewal per Cronjob

Ich hatte ursprünglich einen dauerhaft laufenden certbot-Container im Compose-File – das ist das was man in den meisten Anleitungen sieht. Beim Recherchieren bin ich aber auf ein Problem gestoßen: nginx lädt das erneuerte Zertifikat nicht automatisch. Der Container würde das Zertifikat erneuern, aber nginx würde es erst beim nächsten Neustart laden – im schlimmsten Fall läuft das alte Zertifikat ab ohne dass nginx es merkt.

Die sauberere Lösung ist ein Cronjob auf dem Host der beides in einem Schritt macht:

```bash
crontab -e
```

```
0 3 * * * cd /opt/proxy && docker compose run --rm certbot renew --quiet && docker compose exec nginx nginx -s reload
```

Läuft täglich um 3 Uhr. certbot macht nichts wenn das Zertifikat noch mehr als 30 Tage gültig ist – der Job kann ruhig täglich laufen.

## Deployment mit GitHub Actions

Für CI/CD war mein erster Gedanke `git pull` auf dem Server. Das Problem: das Repo ist privat, `git pull` braucht Git-Zugang, und den für einen Deploy-User einzurichten wäre wieder ein Henne-Ei-Problem – ein SSH-Key für GitHub, ein SSH-Key für den Server, Berechtigungen konfigurieren. Unnötige Komplexität für ein Repo das nur nginx-Configs enthält.

`scp` ist einfacher: GitHub Actions kopiert die Dateien direkt auf den Server. Kein Git auf dem Server, keine Zugangskonfiguration, keine Merge-Konflikte.

```yaml
# .github/workflows/deploy.yml
name: Deploy Proxy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Copy files to server
        uses: appleboy/scp-action@v1
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          source: "."
          target: "/opt/proxy"

      - name: Deploy
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd /opt/proxy
            docker compose up -d
```

Der Deploy-User braucht nur Schreibrechte auf `/opt/proxy` und Zugang zur Docker-Gruppe – kein sudo, kein Git-Zugang.

## Erstes Setup: Reihenfolge

1. Externes Netzwerk anlegen: `docker network create proxy_network`
2. `.env` anlegen mit `CERTBOT_EMAIL`
3. `nginx/conf.d/blog.conf` mit der echten Domain anlegen
4. Sicherstellen dass die Domain bereits auf den Server zeigt – ohne das schlägt die ACME-Challenge fehl
5. `./bootstrap.sh` ausführen
6. Cronjob einrichten
7. GitHub Actions Secrets hinterlegen

Ab dann läuft alles automatisch.
