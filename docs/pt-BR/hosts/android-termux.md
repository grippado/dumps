# Host Android (Termux)

O dumps nasceu num celular Android na sala, rodando Termux. Este guia usa o serviço `termux-runit`, que o `dumps install-dumpsd` cria sozinho. Os passos do Termux em si (pacotes, runit, boot) são práticas do ecossistema e não do código deste repositório; confira a documentação do Termux se algo mudar.

## Preparo no celular

Instale o Termux e o Termux:Boot pelo F-Droid. Não misture fontes de instalação diferentes entre os dois.

```bash
pkg install openssh rsync caddy cloudflared termux-services
```

Inicie o servidor ssh e defina uma senha ou instale sua chave pública em `~/.ssh/authorized_keys`:

```bash
sshd
```

Por padrão o `sshd` do Termux escuta na porta 8022. No notebook, uma entrada em `~/.ssh/config`:

```
Host meucelular
  HostName 192.168.0.50
  Port 8022
  User u0_a123
```

O usuário é o resultado de `whoami` no Termux. O CLI usa ssh sem senha (`BatchMode=yes`), então use chave.

## Manter o celular acordado

Android mata processos em segundo plano. Para o servidor sobreviver:

- Rode `termux-wake-lock` (ou use a ação "Acquire wakelock" da notificação do Termux).
- Desligue a otimização de bateria do Termux, do Termux:Boot e, se usar, do cloudflared (Configurações do Android, Apps, Bateria, Sem restrições).
- Deixe o celular no carregador.

## Serviços runit

O `termux-services` usa o runit. Em sessões ssh, o `SVDIR` não vem definido, e os comandos `sv` falham. Defina-o antes de usar:

```bash
export SVDIR=$PREFIX/var/service
```

O `dumps install-dumpsd` já define o `SVDIR` ao reiniciar o serviço.

Configure no CLI:

```
DUMPS_HOST=meucelular
DUMPS_ROOT=/data/data/com.termux/files/home
DUMPS_URL=https://dumps.example.com
DUMPS_TARGET=linux/arm64
DUMPS_SERVICE=termux-runit
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD do app do admin>
```

Instale o binário e o serviço:

```bash
cli/dumps install-dumpsd
```

O comando grava `$PREFIX/var/service/dumpsd/run` e `log/run` (logs via `svlogd` em `$LOGDIR/sv/dumpsd`). Na primeira vez, ele avisa `serviço dumpsd criado; para ligar: sv-enable dumpsd`. No celular:

```bash
sv-enable dumpsd
```

Nas próximas instalações, o CLI reinicia o serviço sozinho.

Caddy e cloudflared rodam como serviços runit próprios, que você cria. Exemplo para o Caddy, em `$PREFIX/var/service/caddy/run`:

```
#!/data/data/com.termux/files/usr/bin/sh
exec 2>&1
exec caddy run --config /data/data/com.termux/files/home/Caddyfile
```

E para o cloudflared, em `$PREFIX/var/service/cloudflared/run`:

```
#!/data/data/com.termux/files/usr/bin/sh
exec 2>&1
exec cloudflared tunnel run dumps
```

Dê permissão de execução e ative cada um:

```bash
chmod 700 $PREFIX/var/service/caddy/run $PREFIX/var/service/cloudflared/run
```

```bash
sv-enable caddy
```

```bash
sv-enable cloudflared
```

Use o `deploy/Caddyfile` com `/home/you` trocado por `/data/data/com.termux/files/home`, e o `deploy/cloudflared.yml` ([exposure/cloudflare.md](../exposure/cloudflare.md)).

## Subir no boot

Com o Termux:Boot instalado e aberto uma vez, crie `~/.termux/boot/start`:

```
#!/data/data/com.termux/files/usr/bin/sh
termux-wake-lock
sshd
. $PREFIX/etc/profile.d/start-services.sh
```

```bash
chmod +x ~/.termux/boot/start
```

A última linha inicia o supervisor do `termux-services`, que sobe os serviços habilitados.

## Particularidades do binário

O `dumpsd` é um binário estático `linux/arm64` (`CGO_ENABLED=0`). No Android não existem `/etc/resolv.conf` nem CAs nos lugares usuais, então, para buscar as chaves do Cloudflare Access, ele lê direto dos caminhos do Termux:

- DNS: `$PREFIX/etc/resolv.conf` (ou `/data/data/com.termux/files/usr/etc/resolv.conf` e `/etc/resolv.conf`); sem nameserver válido, usa `1.1.1.1` e `8.8.8.8`
- CAs: `$PREFIX/etc/tls/cert.pem`

Isso só importa com `DUMPS_AUTH=access`. Se `$PREFIX` não estiver definido no ambiente do serviço, o binário tenta os caminhos fixos `/data/data/com.termux/files/usr/etc/...` acima.
