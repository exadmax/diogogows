# Instância de demonstração

Diogows separado do de uso real: banco, usuários e arquivos (`desk/`) ficam só nesta pasta (ignorados pelo git; só este README e o
`diogows.config.json` são versionados). Serve para gravar vídeos e apresentar o Diogows sem expor dados pessoais.

```bash
cd server
go run . -config demo/diogows.config.json
```

Abra http://localhost:8091. Na primeira vez o log mostra a senha inicial do `admin` (troca obrigatória no primeiro login).
Para recomeçar do zero, pare o Diogows e apague tudo em `server/demo/` menos `diogows.config.json` e este arquivo.

`central.url` fica vazio de propósito: sem a Central Diogo, câmeras, contatos e agendas da casa aparecem como "indisponível".
Para mostrar a ligação com ela, aponte para uma Central de demonstração (ex.: `"url": "http://127.0.0.1:8080"`).

Sugestão de roteiro: crie 1 ou 2 usuários em Iniciar → Usuários, preencha *Meus arquivos* com documentos, planilhas, slides e
imagens, ponha vídeos em `videos/` e troque mensagens entre os usuários.

Dados antigos (da época em que o Diogows morava dentro da Central): `demo.db` e `desk/` continuam aqui. Para trazê-los, com o
Diogows parado e `demo-diogows.db` ainda inexistente:

```bash
go run . -config demo/diogows.config.json -import-central demo/demo.db
```
