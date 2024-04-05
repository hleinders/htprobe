# htprobe - Web Request Analyzer

Das CLI-Programm **htprobe** dient zur Analyse von Webrequests auf der Kommandozeile. Oft ist es wichtig, schnell einer Redirect-Kette zu folgen, ein Zertifikat zu überprüfen oder die zurück gesendeten Header zu untersuchen. Natürlich können das auch alle modernen Browser mit Hilfe ihrer jeweiligen Entwickler-Tools - aber manchmal ist es auf der Kommandozeile einfach schneller. 

![Screenshot](https://github.com/hleinders/htprobe/blob/main/resources/img/htprobe.png)

## Syntax

Der Aufruf von **htprobe** geschieht so:

``` shell
$ htprobe <subcommand> URL [URL ...][flags]
```

Die optionalen globalen Flags können überall auf der Kommandozeile nach **htprobe** angegeben werden, Flags, die zu einem Subkommando gehören, aber erst nach diesem Kommando. Mit dem Subkommando *help* oder dem Schalter *-h|--help* kann jederzeit eine Hilfe angezeigt werden, mit 

```she
$ htprobe <subcommand> --help
```

die Hilfeseite für ein bestimmtes Subkommando.



Die verfügbaren Module sind:

* **certificate:** Analysiert Server-Zertifikate und zeigt sie an
* **completion:** Erzeugt die Autovervollständigung für die vorgegebene Shell
* **content:** Führt einen Webrequest durch und zeigt den Inhalt an, falls vorhanden.
* **cookies:** Zeigt die Request- und Response-Cookies eines Webrequests
* **headers:** Zeigt die Request- und Response-Header eines Webrequests
* **help:** Zeigt die Hilfe von **htprobe** oder eines Subkommandos an
* **redirects:** Folgt der Redirect-Kette eines Webrequests und zeigt sie an



## Installation

**htprobe** ist in **[Go](https://go.dev/)** geschrieben. man benötigt daher eine funktionierende Entwicklungsumgebung für diese Sprache. Hinweise zur [Installation](https://go.dev/doc/install) und einen Downloadlink für die verschiedenen Plattformen finden sich auf der [Webseite](https://go.dev) von **Go**. 

Sofern alles erfolgreich installiert wurde, sollte dieser Befehl funktionieren:

```shell
$ go version
go version go1.22.2 linux/amd64
```

Danach muss dieses Repository geklont werden und Programm kompiliert werden:

```shell
$ git clone https://github.com/hleinders/htprobe.git
$ cd htprobe
$ go mod tidy
$ go build     # or go install, if GOBIN is set
```



## Beispiele

#### Rufe *nasa.gov* auf und zeige die Redirect-Kette an:

```shell
$ htprobe redirects nasa.gov

URL: http://nasa.gov  [GET]
       ┣━━ (301) ⮕  [GET] https://nasa.gov/
       ┣━━ (302) ⮕  [GET] https://www.nasa.gov/
       ┗━━ (200) ⮕  200 OK

```



#### Untersuche die Header im letzten "Hop":

```shell
$ htprobe headers https://www.nasa.gov/

1:  URL: https://www.nasa.gov/  [GET] (200)
═══════════════════════════════════════════

     Request Header:
     • User-Agent: HtProbe Request Analyzer v1.2 (2024-04-05)

     Response Header:
     • Accept-Ranges: bytes
     • Age: 89
     • Cache-Control: max-age=300, must-revalidate
     • Content-Type: text/html; charset=UTF-8
     • Date: Thu, 04 Apr 2024 09:31:54 GMT
     • Host-Header: a9130478a60e5f9135f765b23f26593b
     • Server: nginx
     • Strict-Transport-Security: max-age=31536000
     • Vary: Accept-Encoding
     • X-Cache: hit
     • X-Launch-Status: Go Flight!
     • X-Rq: hhn1 85 188 443

```



#### Zeige die Zertifikatsinformationen:

```shell
$ htprobe certificate https://www.nasa.gov/

1:  URL: https://www.nasa.gov/  [GET] (200)
═══════════════════════════════════════════

     Certificate(s):
     • CN:          nasa.gov
       SANs:        nasa.gov, www.nasa.gov
       Valid until: 2024-06-25 13:02:10 +0000 UTC
       CA-Chain:    nasa.gov
       ⋘  R3 (Let's Encrypt)

```



## Aliase

Ich persönlich benutze folgende Aliase in meiner Shell:

```shell
alias checkRedirects="htprobe redirects"
alias checkHeader="htprobe headers"
alias checkCert="htprobe certificate"
alias checkCookies="htprobe cookies"
```



Damit kann man die obigen Aufrufe ein wenig vereinfachen, zum Beispiel:

```shell
$ checkRedirects nasa.gov

URL: http://nasa.gov  [GET]
       ┣━━ (301) ⮕  [GET] https://nasa.gov/
       ┣━━ (302) ⮕  [GET] https://www.nasa.gov/
       ┗━━ (200) ⮕  200 OK

$ checkCert https://www.nasa.gov/

1:  URL: https://www.nasa.gov/  [GET] (200)
═══════════════════════════════════════════

     Certificate(s):
     • CN:          nasa.gov
       SANs:        nasa.gov, www.nasa.gov
       Valid until: 2024-06-25 13:02:10 +0000 UTC
       CA-Chain:    nasa.gov
       ⋘  R3 (Let's Encrypt)

```



## Shell Completion

Die meisten modernen Shells (bash, zsh, ...) haben die Fähigkeit, Programme und deren Subkommandos oder Parameter automatisch zu vervollständigen. Den entsprechenden Code erhält man mit

```shell
$ htprobe completion zsh
```

Die kann man automatisieren, z.B. für die *zsh*, in dem man in der $HOME/.zshrc folgende Zeile einfügt:

```shell
source <(htprobe completion zsh)
```



## Danksagung

Dieses Programm wäre ohne die Vorarbeit von so vielen Open Source Programmierern nicht möglich gewesen. Insbesondere möchte ich die folgenden Module nennen: "[cobra](https://pkg.go.dev/github.com/spf13/cobra)" und "[pflag](https://pkg.go.dev/github.com/spf13/pflag)" von Steve Francia (spf13) sowie "[color](https://github.com/fatih/color)" von Fatih Arslan (fatih).

