# Modo local

O modo local não precisa de servidor, ssh, proxy nem túnel. Tudo fica na sua máquina e é servido no loopback. Funciona em macOS e Linux. Com `DUMPS_HOST=local` (o padrão), o CLI copia os arquivos com `rsync` para uma raiz local e roda o `dumpsd` ali.

## Requisitos

- `bash`, `jq`, `rsync` e `python3`
- Go, na versão de `server/go.mod`, para compilar o `dumpsd`
- Opcional, para miniaturas: Chrome ou Chromium, mais `sips` (já vem no macOS) ou ImageMagick (Linux)

## Preparo

O modo local é o padrão, então não precisa de arquivo de configuração. Os padrões são:

| Configuração | Padrão local |
|--------------|--------------|
| Raiz | `${XDG_DATA_HOME:-~/.local/share}/dumps` (`DUMPS_ROOT`) |
| URL pública (origem dos dumps) | `http://localhost:8090` (`DUMPS_URL`) |
| Listener dos dumps | `127.0.0.1:8090` (`DUMPS_LISTEN`) |
| Listener do admin | `127.0.0.1:8091` (`DUMPS_ADMIN_LISTEN`) |
| Origem do admin | `http://localhost:8091`, vinda da porta de `DUMPS_ADMIN_LISTEN` (ou `DUMPS_ADMIN_ORIGIN`, se definida) |
| Página do admin | `http://localhost:8091/_dumps/` (`DUMPS_LIST_URL`) |

Compile e instale o `dumpsd` na raiz. O comando roda os testes de Go antes e compila para a máquina atual.

```bash
cli/dumps install-dumpsd
```

Publique.

```bash
cli/dumps publish ./pagina.html
```

```bash
cli/dumps publish ./minha-pasta --title "Minha pasta" --id minha-pasta
```

Sirva os dumps e o admin no loopback. O comando refaz a home e a listagem e executa `dumpsd serve --auth none --static` com os dois listeners. `Ctrl-C` encerra.

```bash
cli/dumps serve-local
```

Navegue:

- http://localhost:8090 para a home
- http://localhost:8090/<id>/ para um dump
- http://localhost:8091/_dumps/ para a listagem do admin (arquivar, renomear, apagar)

Abra o admin por `localhost`, não por `127.0.0.1`. A API compara o header `Origin` como texto exato, então em `http://127.0.0.1:8091` as ações do admin falham com 403.

Gerencie pelo terminal:

```bash
cli/dumps list
```

```bash
cli/dumps archive <id>
```

Veja a configuração efetiva:

```bash
cli/dumps config
```

## Duas origens

O `dumps serve-local` inicia um único processo `dumpsd` com dois listeners, então os dumps e o admin são origens diferentes (portas diferentes):

| Origem | Endereço | O que serve |
|--------|----------|-------------|
| Dumps | `DUMPS_LISTEN` (padrão `127.0.0.1:8090`) | Só `dumps/`: a home, cada dump e as miniaturas. `/_dumps` devolve 404 e caminhos com ponto no início devolvem 404 |
| Admin | `DUMPS_ADMIN_LISTEN` (padrão `127.0.0.1:8091`) | `/` redireciona para `/_dumps/`, que é a listagem. A API fica em `/_dumps/api/*`. O resto devolve 404 |

A API só aceita a origem do admin: o header `Origin` de todo `POST` precisa ser igual à origem do admin, caractere a caractere. A origem do admin é `DUMPS_ADMIN_ORIGIN` ou, se estiver vazia, `http://localhost:<porta de DUMPS_ADMIN_LISTEN>`. O JavaScript de um dump roda na origem dos dumps, que não tem API, e o `Origin` dele não confere, então um dump não consegue arquivar, renomear nem apagar nada.

Regras que o `dumpsd serve --static` impõe:

- Só funciona com `--auth none`.
- Exige `--admin-listen`, numa porta diferente de `--listen`.
- Os dois endereços precisam ser loopback (`127.0.0.1:`, `localhost:` ou `[::1]:`).

Para usar outras portas, mude os três valores juntos. Como `DUMPS_ADMIN_ORIGIN` está vazia, a origem do admin acompanha a nova porta do admin.

```bash
DUMPS_URL=http://localhost:9000 DUMPS_LISTEN=127.0.0.1:9000 DUMPS_ADMIN_LISTEN=127.0.0.1:9001 cli/dumps serve-local
```

Se `DUMPS_URL` e `DUMPS_LISTEN` divergirem, os links quebram. A checagem de `Origin` quebra quando `DUMPS_ADMIN_ORIGIN` é definida com algo diferente da origem por onde você abre o admin.

## Onde ficam os dados

A raiz padrão é `~/.local/share/dumps` (ou `$XDG_DATA_HOME/dumps`). Mude com `DUMPS_ROOT`. O layout está descrito em [architecture.md](architecture.md).

## Miniaturas

O `dumps publish` tira um screenshot de 1280x800 do `index.html` com o Chrome headless e reduz para um JPEG de 640 px de largura.

- Ordem de busca do Chrome: `DUMPS_CHROME`, depois Google Chrome, Brave e Chromium em `/Applications` (macOS), depois `google-chrome`, `chromium` e `chromium-browser` no `PATH`.
- Redução: `sips` no macOS, `magick` ou `convert` (ImageMagick) no Linux.

Sem navegador ou sem conversor, a publicação segue. O CLI avisa e a home mostra um placeholder com a inicial do título. As tags Open Graph entram sem imagem.

A página é aberta por `file://`, então referências a caminhos absolutos (`/assets/x.css`) não carregam na miniatura. Use caminhos relativos nas pastas que você publica.

## Observações

- `--auth none` aceita só conexões diretas do loopback. Também recusa qualquer pedido que traga `X-Forwarded-For`, `Forwarded`, `Cf-Connecting-Ip`, `X-Real-Ip` ou `Cf-Access-Jwt-Assertion`. O `dumpsd serve` recusa escutar em endereço que não seja loopback. Veja [auth.md](auth.md).
- `--auth none` é só para o `serve-local`. O CLI recusa `DUMPS_AUTH=none` ao instalar um serviço.
- O JavaScript de um dump não consegue chamar a API do admin. Os dumps ainda dividem uma única origem entre si, então publique só HTML em que você confia. Veja [security.md](security.md).
- Com o padrão `DUMPS_SERVICE=none`, nenhum serviço é instalado. Rode `dumps serve-local` quando precisar.
