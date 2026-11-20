# Modo local

Publique e navegue no seu próprio computador (macOS ou Linux), sem servidor, túnel nem login. Com `DUMPS_HOST=local` (o padrão), o CLI copia os arquivos com `rsync` para a raiz local e roda o `dumpsd` ali.

## Requisitos

- bash, `jq`, `rsync` e `python3`
- Go (a versão de `server/go.mod`), só para compilar o `dumpsd`
- Opcional: Chrome ou Chromium para a miniatura (veja abaixo)

## Passo a passo

```bash
cli/dumps install-dumpsd
```

O comando roda `go test`, compila o `dumpsd` para a sua máquina e instala em `<raiz>/bin/dumpsd`. Como `DUMPS_SERVICE` é `none`, nenhum serviço é criado.

```bash
cli/dumps publish ./pagina.html
```

O CLI imprime a URL do dump (`http://localhost:8090/<id>/`).

```bash
cli/dumps serve-local
```

Abra http://localhost:8090 para a home e http://localhost:8090/_dumps/ para a listagem com arquivar, renomear e apagar. `Ctrl-C` encerra.

## Onde ficam os dados

A raiz padrão é `~/.local/share/dumps` (ou `$XDG_DATA_HOME/dumps`). Mude com `DUMPS_ROOT`. O layout é o descrito em [architecture.md](architecture.md).

## O que o serve-local faz

Refaz a home e a listagem e executa `dumpsd serve --auth none --static` no endereço de `DUMPS_LISTEN` (padrão `127.0.0.1:8090`). O modo `--auth none` só aceita conexões do loopback. A origem aceita nos POSTs é `DUMPS_ADMIN_ORIGIN` ou, se vazia, `DUMPS_URL`.

Para usar outra porta, mude `DUMPS_LISTEN` e `DUMPS_URL` juntos (por exemplo `127.0.0.1:9000` e `http://localhost:9000`). Se os dois divergirem, os links e a checagem de `Origin` quebram.

A checagem de `Origin` é uma comparação exata de texto: abrir a listagem por `http://127.0.0.1:8090` em vez de `http://localhost:8090` faz as ações do admin falharem com 403.

Os arquivos são servidos com URLs limpas (`/<id>/`, sem `.html`), como no Caddy de produção.

## Miniaturas

A miniatura é um screenshot de 1280x800 do `index.html`, reduzido para 640 px de largura em JPEG. Precisa de duas peças:

1. Um navegador Chromium. O CLI procura, nesta ordem: `DUMPS_CHROME`, Google Chrome, Brave e Chromium em `/Applications` (macOS), e depois `google-chrome`, `chromium` e `chromium-browser` no `PATH`.
2. Um conversor de imagem: `sips` (já vem no macOS) ou ImageMagick (`magick` ou `convert`) no Linux.

Sem alguma delas, o publish avisa `miniatura não gerada, a home usa um placeholder` e segue. O placeholder é a inicial do título.

A página é aberta por `file://`, então referências a caminhos absolutos (`/assets/x.css`) não carregam na miniatura. Use caminhos relativos nas pastas que você publica.
