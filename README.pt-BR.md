# dumps

Um lugar auto-hospedado para publicar artefatos HTML estáticos (um `.html` ou uma pasta com `index.html`) em `https://<host>/<id>/`.
Nasceu num celular Android na sala (Termux + Caddy + Cloudflare Tunnel), mas roda em qualquer Linux, macOS ou Android, ou totalmente local.

[English](README.md) | Português

## Recursos

- Publicação de um arquivo `.html` ou de uma pasta com `index.html`, em `https://<host>/<id>/`.
- Home pública com cards e miniatura, em quadro ou lista, com links em destaque.
- Barra no topo de cada dump com compartilhar (WhatsApp, Telegram, copiar link) e modo limpo (tecla Esc).
- Rodapé com autores e prévia Open Graph geradas na publicação.
- Google Analytics 4 opcional (`DUMPS_GA_ID`).
- Listagem privada de admin: arquivar, desarquivar, renomear e apagar, protegida por login (Cloudflare Access ou Tailscale).
- Ativação atômica com rollback: uma publicação que falha não deixa o dump pela metade no ar.
- Modo local, sem servidor, sem túnel e sem login.

Os textos da interface (home, barra, rodapé e listagem) hoje estão em português. A internacionalização (i18n) está planejada.

## Arquitetura

```
 Notebook do autor                       Servidor (Raspberry Pi, VPS, NAS, Mac, celular)
+--------------------+  ssh / rsync   +---------------------------------------------+
| cli/dumps          | -------------> | raiz (DUMPS_ROOT)                           |
| Chrome headless    |                |   dumps/          conteúdo público + home   |
| (miniatura)        |                |   dumps-meta/     manifests dos dumps       |
+--------------------+                |   dumps-staging/  área de preparo           |
                                      |   bin/dumpsd      injeta, ativa, renderiza  |
                                      |   admin/          pasta da listagem         |
                                      +---------------------------------------------+
                                           |                         |
                          Caddy :8081 (público)       Caddy :8080 (admin) + dumpsd :8090 (API)
                                           |                         |
                                           +---------- túnel --------+
                                                       |
                                                    internet
                                     (o host do admin fica atrás de login)
```

Detalhes em [docs/pt-BR/architecture.md](docs/pt-BR/architecture.md).

## Início rápido (modo local)

Você precisa de bash, [Go](https://go.dev/dl/) (a versão de `server/go.mod`), `jq`, `rsync` e `python3`.

```bash
git clone https://github.com/grippado/dumps.git
```

```bash
cd dumps
```

```bash
cli/dumps install-dumpsd
```

```bash
cli/dumps publish ./pagina.html
```

```bash
cli/dumps serve-local
```

Abra http://localhost:8090. Sem configuração, o CLI usa `DUMPS_HOST=local` e guarda tudo em `~/.local/share/dumps`.
Para miniaturas, instale Chrome ou Chromium (veja [docs/pt-BR/local.md](docs/pt-BR/local.md)).

## Documentação

- [Arquitetura](docs/pt-BR/architecture.md)
- [Configuração](docs/pt-BR/configuration.md)
- [Modo local](docs/pt-BR/local.md)
- Hosts: [Linux](docs/pt-BR/hosts/linux.md), [macOS](docs/pt-BR/hosts/macos.md), [Android (Termux)](docs/pt-BR/hosts/android-termux.md)
- Exposição: [Cloudflare Tunnel e Access](docs/pt-BR/exposure/cloudflare.md), [alternativas](docs/pt-BR/exposure/alternatives.md)
- [Autenticação do admin](docs/pt-BR/auth.md)
- [Segurança](docs/pt-BR/security.md)

## Desenvolvimento

```bash
make check
```

`make check` roda gofmt, `go vet`, shellcheck e os testes Go, o mesmo que o CI.

## Licença

[MIT](LICENSE).
