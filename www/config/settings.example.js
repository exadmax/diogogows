// Padrões do Diogows. Para personalizar, crie config/settings.js (ignorado pelo git) com:
//   window.CD_SETTINGS_LOCAL = { central: { port: 8080 } };   (8080 é a porta da Central no seu PC de desenvolvimento)  (só o que mudar; é mesclado sobre este arquivo)
window.CD_SETTINGS = {
  // Central Diogo (o painel da casa), que é outro programa. Os atalhos "Painel e serviços da casa" do menu Iniciar e o botão
  // "Voltar ao painel da casa" abrem este endereço. Use "port" (mesma máquina) ou "url" completa (outra máquina ou outro domínio).
  central: { port: 80 },

  // Atalhos usados enquanto a Central não responde. Com ela no ar, vale o menu salvo na Config dela (/api/settings/menu).
  // {host} = http://endereço-da-máquina; um "path" sem "port" é uma página da Central.
  links: [
    { label: 'Jogos', sub: 'WebRCade', icon: 'gamepad', port: 8080, path: '/jogos/play/?feed={host}:8080/roms/feed.json' },
    { label: 'Geladeira', sub: 'Itens e validades', icon: 'fridge', path: '/estoque.html' },
    { label: 'Serviços', sub: 'Profissionais', icon: 'phone', path: '/contatos.html' },
    { label: 'Câmeras', sub: 'Ao vivo', icon: 'camera', path: '/camaras.html' },
    { label: 'Música', sub: 'Navidrome', icon: 'music', port: 4533, path: '/' },
    { label: 'Vídeos', sub: 'Abrir no VLC', icon: 'film', path: '/videos.html' },
    { label: 'Livros', sub: 'PDF e EPUB', icon: 'book', path: '/livros.html' },
    { label: 'Config', sub: 'Servidor', icon: 'gear', path: '/config.html' }
  ],

  // LiteOffice Completo (LibreOffice no navegador): de onde baixar os arquivos do ZetaOffice (soffice.js, .wasm, .data; ~50 MB).
  // Padrão: o CDN do ZetaOffice. Para servir de dentro do Diogows, copie os arquivos para www/vendor/zetaoffice/ e use base: 'vendor/zetaoffice/'.
  zetaoffice: { base: 'https://cdn.zetaoffice.net/zetaoffice_latest/' }
};
