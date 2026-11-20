# Host Android (Termux)

O dumps nasceu aqui: um celular Android no wifi de casa, servindo por um túnel. O celular roda o `dumpsd`, o Caddy e o túnel como serviços runit. O CLI roda no seu notebook. Este guia usa o serviço `termux-runit`, que o `dumps install-dumpsd` cria sozinho.

Os passos do Termux em si (pacotes, runit, boot) seguem o ecossistema do Termux e não fazem parte do código deste repositório, então confira a documentação do Termux se algo mudar.

## 1. Instalar o Termux

Instale o Termux e o Termux:Boot pelo F-Droid, não pela Play Store (aquela versão está desatualizada). Não misture fontes de instalação entre os dois. Depois, no Termux:

```bash
pkg install openssh rsync caddy cloudflared termux-services
```

Reinicie o Termux uma vez para o `termux-services` configurar o runit.

## 2. ssh para o celular

Inicie o sshd no Termux (ele escuta na porta 8022 por padrão) e adicione a chave pública do notebook em `~/.ssh/authorized_keys`.

```bash
sshd
```

```bash
passwd
```

No notebook, em `~/.ssh/config`:

```text
Host celular
  HostName 192.168.0.50
  Port 8022
  User u0_a123
```

Use o endereço real e o usuário que o `whoami` mostra no Termux. Use chave, porque o CLI roda o ssh com `BatchMode=yes`.

## 3. Configuração do CLI

```text
DUMPS_HOST=celular
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_TARGET=linux/arm64
DUMPS_SERVICE=termux-runit
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<tag AUD>
```

A raiz padrão é a home do Termux, então os dumps ficam em `~/dumps`. Para fixá-la, defina `DUMPS_ROOT=/data/data/com.termux/files/home`.

Para um serviço, `DUMPS_AUTH` é `access` ou `tailscale`. `none` é recusado. Os valores que vão para o serviço só podem ter letras, números e `: / . _ @ -`.

`DUMPS_TARGET=linux/arm64` gera um binário estático (`CGO_ENABLED=0`) que roda no Termux. Dentro do Termux não existem `/etc/resolv.conf` nem CAs nos lugares usuais, então, para buscar as chaves do Cloudflare Access, o `dumpsd` lê:

- DNS: `$PREFIX/etc/resolv.conf` (ou `/data/data/com.termux/files/usr/etc/resolv.conf` e `/etc/resolv.conf`). Sem nameserver válido, usa `1.1.1.1` e `8.8.8.8`.
- Certificados CA: `$PREFIX/etc/tls/cert.pem` (ou `/data/data/com.termux/files/usr/etc/tls/cert.pem`).

Isso só importa com `DUMPS_AUTH=access`.

## 4. Instalar o dumpsd e o serviço

```bash
cli/dumps install-dumpsd
```

Com `DUMPS_SERVICE=termux-runit`, o comando cria `$PREFIX/var/service/dumpsd/` com um script `run` e um `log/run` (logs via `svlogd` em `$LOGDIR/sv/dumpsd`). Se o serviço nunca foi habilitado, o CLI imprime `serviço dumpsd criado; para ligar: sv-enable dumpsd`. No celular:

```bash
sv-enable dumpsd
```

Quando o serviço já está habilitado, o `install-dumpsd` o reinicia por você.

## 5. Caddy e o túnel como serviços

Rode o Caddy (com o `deploy/Caddyfile`, trocando `/home/you` por `/data/data/com.termux/files/home`) e o `cloudflared` (com o `deploy/cloudflared.yml`) como serviços runit próprios, e habilite os dois. Cada serviço precisa de uma pasta em `$PREFIX/var/service/<nome>/` com um script `run` executável que faça `exec` do programa.

Exemplo para o Caddy, em `$PREFIX/var/service/caddy/run`:

```text
#!/data/data/com.termux/files/usr/bin/sh
exec 2>&1
exec caddy run --config /data/data/com.termux/files/home/Caddyfile
```

Exemplo para o cloudflared, em `$PREFIX/var/service/cloudflared/run`:

```text
#!/data/data/com.termux/files/usr/bin/sh
exec 2>&1
exec cloudflared tunnel run dumps
```

```bash
chmod 700 $PREFIX/var/service/caddy/run $PREFIX/var/service/cloudflared/run
```

```bash
sv-enable caddy
```

```bash
sv-enable cloudflared
```

A configuração do túnel está em [../exposure/cloudflare.md](../exposure/cloudflare.md).

## 6. Usar o sv por ssh

O `sv` encontra os serviços pelo `SVDIR`, que o Termux só define em sessões interativas. Por ssh, defina você mesmo:

```bash
SVDIR=$PREFIX/var/service sv status dumpsd
```

O `dumps install-dumpsd` já define o `SVDIR` ao reiniciar o serviço.

## 7. Manter o celular vivo

O Android mata aplicativos em segundo plano com agressividade. Faça tudo isto:

- Segure um wake lock de dentro do Termux:

```bash
termux-wake-lock
```

- Desligue a otimização de bateria do Termux (e do Termux:Boot) nas configurações do Android e permita que rodem em segundo plano.
- Coloque um script em `~/.termux/boot/` (o Termux:Boot o executa depois de um reboot) que inicie o `sshd`, pegue o wake lock e suba os serviços runit (veja a documentação do `termux-services` para o comando).
- Deixe o celular no carregador, no wifi, e evite os modos de economia de bateria.

## 8. Publicar

```bash
cli/dumps publish ./pagina.html
```

As miniaturas são geradas no seu notebook, então o celular só guarda e serve os arquivos.
