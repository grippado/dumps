# dumps

Um lugar auto-hospedado para publicar artefatos HTML estáticos. Publique um único arquivo `.html` ou uma pasta com `index.html` e receba um link em `https://<host>/<id>/` para compartilhar.

Nasceu num celular Android na sala (Termux, Caddy e um Cloudflare Tunnel), mas roda em qualquer máquina Linux, macOS ou Android, ou totalmente local no seu notebook.

[English](README.md) | Português

Página do projeto: https://grippado.github.io/dumps/

## Só quer artefatos melhores? Use a skill

`dump-this` é uma skill que monta a página do jeito que um especialista montaria (um analista
para relatórios, um editor para histórias, um coach de apresentação para slides, um especialista
em visualização de dados para dashboards, e mais). Funciona no seu harness favorito, com ou sem
um servidor dumps:

- **Claude Code:** descompacte o [dump-this.zip](https://grippado.github.io/dumps/dump-this.zip)
  em `~/.claude/skills` (sem clonar nada) e digite `/dump-this <o que você quer>`. Codex: o mesmo
  zip em `~/.codex/skills`. Comandos em [docs/pt-BR/agent.md](docs/pt-BR/agent.md).
- **claude.ai:** suba o [dump-this.zip](https://grippado.github.io/dumps/dump-this.zip) em
  Configurações, Capacidades, Skills, e receba a página como artefato.
- **ChatGPT:** cole o [dump-this-chatgpt.txt](https://grippado.github.io/dumps/dump-this-chatgpt.txt)
  num GPT personalizado ou num Projeto, e receba a página no canvas.

Ela é opinativa de propósito: um dial de estilo lido das suas próprias palavras vai de "sóbrio" a
"inovação máxima", e ela recusa o visual padrão de artefato. Veja um exemplo de cada tipo em
[grippado.github.io/dumps/examples](https://grippado.github.io/dumps/examples/).

Guia completo: [docs/pt-BR/agent.md](docs/pt-BR/agent.md).

## Recursos

- Publicação de um arquivo `.html` ou de uma pasta com `index.html`; cada dump recebe um id aleatório curto (ou o que você escolher com `--id`).
- Home pública com cards e miniaturas, alternância entre quadro e lista e links em destaque opcionais.
- Barra no topo de cada dump com botão de compartilhar (WhatsApp, Telegram, copiar link, compartilhamento nativo em telas de toque) e modo limpo (tecla Esc) que esconde a barra e o rodapé.
- Rodapé com a URL do dump e os autores.
- Tags Open Graph e Twitter card injetadas em cada arquivo HTML publicado, com a miniatura como imagem de prévia.
- Google Analytics 4 opcional (`DUMPS_GA_ID`).
- Listagem privada de admin atrás de login (Cloudflare Access ou Tailscale): arquivar, desarquivar, renomear e apagar.
- Publicação atômica com rollback: uma publicação que falha deixa a versão anterior no ar.
- Miniaturas geradas localmente com Chrome ou Chromium headless.
- Uma skill de agente, a `dump-this`, que deixa o Claude Code e o Codex transformarem um pedido numa página e publicá-la escondida para você revisar.
- Modo local sem servidor, sem túnel e sem login, com os dumps e o admin em origens separadas.
- Um binário Go pequeno no servidor (`dumpsd`, sem CGO) e um CLI em bash na sua máquina.

Os textos da interface (home, barra, rodapé, admin e mensagens do CLI) hoje estão em português. A internacionalização (i18n) está planejada.

## Arquitetura

```text
notebook                       servidor (qualquer máquina Linux, macOS ou Android)
--------                       ----------------------------------------------------
cli/dumps  --- ssh/rsync --->  DUMPS_ROOT/
 - gera miniaturas               dumps/          arquivos públicos, uma pasta por id
 - envia o HTML cru              dumps-meta/     um manifest JSON por dump, .site.json
                                 dumps-staging/  envios à espera de ativação
                                 bin/dumpsd      prepare, activate, render, apply, serve
                                 admin/          listagem privada (listing_dir)

                                 Caddy :8081 (público)  serve só dumps/
                                 Caddy :8080 (admin)    listagem + proxy para o dumpsd :8090
                                          |
                                 túnel (cloudflared, Tailscale, ...)
                                          |
                                      internet
                                 host público: dumps.example.com
                                 host do admin: admin.example.com (atrás de login)
```

No modo local (`DUMPS_HOST=local`) não há ssh, Caddy nem túnel: o `dumps serve-local` faz o `dumpsd` servir os dumps em `127.0.0.1:8090` e o admin em `127.0.0.1:8091`, duas origens no loopback. Detalhes em [docs/pt-BR/architecture.md](docs/pt-BR/architecture.md).

## Início rápido (modo local)

O modo local publica e serve na sua própria máquina. Você precisa de bash, jq, rsync, python3 e Go (a versão de `server/go.mod`). Chrome ou Chromium headless é opcional e só serve para as miniaturas.

Clone o repositório:

```bash
git clone https://github.com/grippado/dumps.git
```

```bash
cd dumps
```

Compile e instale o `dumpsd` na raiz local (`~/.local/share/dumps`). Não precisa rodar `make build`.

```bash
cli/dumps install-dumpsd
```

Publique uma página. O comando imprime a URL dela.

```bash
cli/dumps publish ./pagina.html
```

Inicie o servidor local:

```bash
cli/dumps serve-local
```

Abra duas URLs:

- http://localhost:8090 para a home e os dumps
- http://localhost:8091/_dumps/ para a listagem do admin (arquivar, renomear, apagar)

Use `localhost`, não `127.0.0.1`: a API do admin só aceita a origem exata `http://localhost:8091`.

Opcional, ainda dentro da pasta do repo: instale a skill de agente e o comando `/dump-this` para o Claude Code e o Codex.

```bash
cli/dumps install-agent all
```

O agente executa o comando `dumps`, então ponha `cli/dumps` no seu `PATH`. Mais em [docs/pt-BR/local.md](docs/pt-BR/local.md) e [docs/pt-BR/agent.md](docs/pt-BR/agent.md).

## Documentação

- [Arquitetura](docs/pt-BR/architecture.md): componentes, layout de diretórios, fluxo do publish, códigos de saída
- [Configuração](docs/pt-BR/configuration.md): todas as chaves `DUMPS_*`, o `site.json` e os subcomandos do `dumpsd`
- [Modo local](docs/pt-BR/local.md)
- [Skill de agente](docs/pt-BR/agent.md): `dump-this`, as receitas e o fluxo de revisão
- Hosts: [Linux](docs/pt-BR/hosts/linux.md), [macOS](docs/pt-BR/hosts/macos.md), [Android (Termux)](docs/pt-BR/hosts/android-termux.md)
- Exposição: [Cloudflare Tunnel e Access](docs/pt-BR/exposure/cloudflare.md), [alternativas](docs/pt-BR/exposure/alternatives.md)
- [Autenticação do admin](docs/pt-BR/auth.md)
- [Segurança](docs/pt-BR/security.md)
- [Contribuindo](CONTRIBUTING.md) e [política de segurança](SECURITY.md) (em inglês)

## Desenvolvimento

```bash
make check
```

O `make check` roda gofmt, `go vet`, shellcheck e os testes de Go, o mesmo que o CI.

## Feito com flux

O dumps foi feito de ponta a ponta com o [flux](https://grippado.github.io/flux/), uma família de
comandos para agentes de IA: cada fatia foi refinada, construída, revisada e iterada com ele, e a
história da construção é ela mesma um dump em
[dumps.gripp.link/historia](https://dumps.gripp.link/historia/).

## Licença

MIT, veja [LICENSE](LICENSE).
