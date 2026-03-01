# Server Hardening: nftables, Docker und fail2ban

**Kategorie:** Infrastruktur

Dieser Artikel dokumentiert wie ich den Server abgesichert habe – und warum der Weg dorthin länger war als erwartet. Das Zusammenspiel von nftables und Docker hat mich mehrere Stunden gekostet und ist ein Thema das in den meisten Anleitungen entweder falsch oder gar nicht erklärt wird.

## SSH Hardening

Das Erste was ich nach der Installation gemacht habe: Passwort-Authentifizierung für SSH deaktivieren. Nur Public-Key.

In `/etc/ssh/sshd_config`:

```
PasswordAuthentication no
PubkeyAuthentication yes
```

```bash
sudo systemctl restart sshd
```

Kein besonderer Grund den Port zu ändern – Security through obscurity bringt wenig, fail2ban erledigt den Rest.

## nftables

Ich nutze nftables statt iptables. Auf Arch ist nftables der Standard, und ich wollte die Firewall vollständig verstehen statt eine fertige Lösung zu kopieren.

Die Config liegt in `/etc/nftables.conf`. Meine Grundregel: alles droppen was nicht explizit erlaubt ist.

```
#!/usr/sbin/nft -f
flush ruleset

table inet filter {
    chain input {
        type filter hook input priority 0;
        policy drop;

        iif lo accept
        ct state established,related accept

        ip protocol icmp accept
        ip6 nexthdr icmpv6 accept

        tcp dport 22 accept
        tcp dport 80 accept
        tcp dport 443 accept
    }

    chain output {
        type filter hook output priority 0;
        policy accept;
    }
}
```

Keine `forward`-Chain – dazu gleich mehr.

Neu laden:

```bash
sudo nft -f /etc/nftables.conf
sudo systemctl enable nftables
```

## Docker und nftables – wo es kompliziert wird

Hier war mein ursprünglicher Plan: Docker so konfigurieren dass es keine eigenen Firewall-Regeln schreibt, und alles selbst über nftables kontrollieren. In `/etc/docker/daemon.json`:

```json
{
  "iptables": false
}
```

Das klingt sauber – eine Firewall, volle Kontrolle. In der Praxis funktioniert es nicht.

### Der erste Versuch: manuelles Masquerading

Ohne Docker's iptables-Regeln haben Container keinen Internetzugang. Ich habe versucht das manuell über nftables zu lösen – eine `forward`-Chain und eine NAT-Masquerade-Regel:

```
table inet filter {
    chain forward {
        type filter hook forward priority 0;
        policy drop;

        ct state established,related accept
        iifname "br-+" oifname != "br-+" accept
        oifname "br-+" iifname != "br-+" accept
    }
}

table ip nat {
    chain postrouting {
        type nat hook postrouting priority 100;
        oifname "ens3" masquerade
    }
}
```

Das Problem: `br-+` als Wildcard funktioniert in nftables nur mit `iifname` (Laufzeit-Match), nicht mit `iif` (wird beim Laden validiert). Mit `iifname` laden die Regeln zwar, aber die Container kamen trotzdem nicht ins Internet.

### Debugging

Stundenlange Fehlersuche. tcpdump zeigte dass Pakete die Bridge verlassen aber nie auf `ens3` ankamen. Der NAT-Counter blieb bei 0. Ich habe versucht:

- Konkrete Bridge-Interface-Namen statt Wildcards
- `br_netfilter` Modul laden
- `net.bridge.bridge-nf-call-iptables=1` setzen
- Die `forward`-Chain komplett rausnehmen

Nichts half. Erst als ich die Firewall komplett deaktiviert habe (`sudo nft flush ruleset`) und der Ping immer noch nicht funktionierte, war klar: das Problem liegt nicht an den Firewall-Regeln.

### Die eigentliche Ursache

`iptables: false` in der Docker-Config deaktiviert nicht nur das Schreiben von Regeln – es deaktiviert auch das interne NAT das Docker für Container-Internetverbindungen einrichtet. Ohne das kommt kein Container ins Internet, egal wie die Firewall konfiguriert ist.

### Die Lösung: iptables-nft

Der saubere Weg ist nicht `iptables: false`, sondern das iptables-Backend auf nftables umstellen. Auf Arch:

```bash
sudo ln -sf /usr/bin/iptables-nft /usr/bin/iptables
sudo ln -sf /usr/bin/ip6tables-nft /usr/bin/ip6tables
```

Prüfen:

```bash
iptables --version
# iptables v1.8.x (nf_tables)  <- das wollen wir sehen
```

Dann Docker neu starten:

```bash
sudo systemctl restart docker
```

Jetzt schreibt Docker seine Regeln über die iptables-Kompatibilitätsschicht in nftables. Das sieht man in `nft list ruleset`:

```
# Warning: table ip nat is managed by iptables-nft, do not touch!
table ip nat {
    chain POSTROUTING {
        ip saddr 172.20.0.0/16 oifname != "br-e528d982046f" xt target "MASQUERADE"
        ...
    }
}
```

Die Aufteilung ist sauber: `table inet filter` gehört mir, `table ip filter` und `table ip nat` gehören Docker. Keine zwei separaten Firewall-Ebenen, kein Konflikt.

Die `forward`-Chain in meiner Config ist komplett raus – Docker verwaltet das selbst und macht es richtig.

### `/etc/docker/daemon.json` anpassen

```json
{}
```

Oder die Datei komplett löschen. `iptables: false` ist weg.

### Legacy-Tabellen aufräumen

Nach dem Wechsel auf iptables-nft ist ein Schritt wichtig der in den meisten Anleitungen fehlt: die alten iptables-legacy-Tabellen aufräumen.

Das Problem: beide Systeme laufen parallel im Kernel in getrennten Tabellen. Docker hatte beim letzten Start noch mit iptables-legacy Regeln geschrieben – darunter eine FORWARD-Chain mit `policy DROP`. Nach dem Wechsel auf iptables-nft schreibt Docker neue Regeln in die nft-Tabellen, aber die alten legacy-Regeln bleiben. Die legacy FORWARD-Chain blockiert dann allen Container-zu-Container Traffic, egal was die nft-Regeln erlauben.

```bash
sudo iptables-legacy -P FORWARD ACCEPT
sudo iptables-legacy -F
sudo iptables-legacy -X
sudo iptables-legacy -t nat -F
sudo iptables-legacy -t nat -X
sudo iptables-legacy -t mangle -F
sudo iptables-legacy -t mangle -X
```

Danach Docker neu starten:

```bash
sudo systemctl restart docker
```

Beim nächsten Serverneustart sind die legacy-Regeln automatisch weg – Docker schreibt dann nur noch in nft. Das Aufräumen ist nur einmalig nach dem Wechsel nötig.

## br_netfilter

Damit die iptables-nft Regeln auf Bridge-Traffic greifen, muss das Kernelmodul `br_netfilter` geladen sein:

```bash
sudo modprobe br_netfilter
```

Dauerhaft in `/etc/modules-load.d/br_netfilter.conf`:

```
br_netfilter
```

Und die zugehörigen sysctl-Werte in `/etc/sysctl.d/99-docker.conf`:

```
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
net.ipv4.ip_forward=1
```

```bash
sudo sysctl -p /etc/sysctl.d/99-docker.conf
```

## fail2ban

fail2ban überwacht SSH und sperrt IPs nach zu vielen fehlgeschlagenen Versuchen. Auf Arch:

```bash
sudo pacman -S fail2ban
```

Die Konfiguration landet in `/etc/fail2ban/jail.local`. Wichtig ist `banaction = nftables-multiport` – damit schreibt fail2ban seine Sperr-Regeln direkt in nftables statt in iptables:

```ini
[DEFAULT]
bantime  = 3600
findtime = 600
maxretry = 3
backend  = systemd
banaction = nftables-multiport
ignoreip = 127.0.0.1/8 ::1

[sshd]
enabled = true
port = 22
```

fail2ban verwaltet eine eigene nftables-Table die automatisch angelegt wird:

```
table inet f2b-table {
    set addr-set-sshd {
        type ipv4_addr
        elements = { 1.2.3.4, ... }
    }

    chain f2b-chain {
        type filter hook input priority filter - 1; policy accept;
        tcp dport 22 ip saddr @addr-set-sshd reject with icmp port-unreachable
    }
}
```

```bash
sudo systemctl enable --now fail2ban
sudo fail2ban-client status sshd
```

## Das Ergebnis

Die finale `/etc/nftables.conf` ist deutlich schlanker als mein erster Versuch:

```
#!/usr/sbin/nft -f
flush ruleset

table inet filter {
    chain input {
        type filter hook input priority 0;
        policy drop;

        iif lo accept
        ct state established,related accept

        ip protocol icmp accept
        ip6 nexthdr icmpv6 accept

        tcp dport 22 accept
        tcp dport 80 accept
        tcp dport 443 accept
    }

    chain output {
        type filter hook output priority 0;
        policy accept;
    }
}
```

Keine `forward`-Chain, keine NAT-Regeln – Docker erledigt das über iptables-nft. fail2ban verwaltet seine eigene Table. Ich kontrolliere nur was ich kontrollieren muss.

Der wichtigste Lerneffekt: `iptables: false` in Docker ist eine Falle. Es klingt nach mehr Kontrolle, ist aber in der Praxis nicht sinnvoll nutzbar ohne das NAT komplett selbst zu reimplementieren – und das ist fehleranfälliger als Docker es zu überlassen.
