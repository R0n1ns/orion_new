```html
<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>Пет-проект: Чат на WebSockets</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f4f4f4;
      color: #333;
      line-height: 1.6;
      padding: 20px;
    }
    .container {
      max-width: 800px;
      margin: 0 auto;
      background: #fff;
      padding: 20px;
      box-shadow: 0 0 10px rgba(0,0,0,0.1);
    }
    h1, h2 {
      text-align: center;
    }
    ul {
      list-style-type: disc;
      margin-left: 20px;
    }
    .todo {
      color: #b12704;
    }
    a {
      color: #0066cc;
      text-decoration: none;
    }
    a:hover {
      text-decoration: underline;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Пет-проект: Чат на WebSockets</h1>
    <p>Этот проект представляет собой современное веб-приложение для обмена сообщениями в реальном времени, реализованное с использованием новейших технологий как на стороне сервера, так и на стороне клиента.</p>
    
    <h2>Технологический стек</h2>
    <ul>
      <li><strong>Бэкенд:</strong> Go с поддержкой WebSockets (ws) для эффективного обмена данными.</li>
      <li><strong>Фронтенд:</strong> HTML, CSS и JavaScript для создания удобного и интуитивного интерфейса.</li>
      <li><strong>Формат обмена данными:</strong> JSON для передачи сообщений между клиентом и сервером.</li>
    </ul>
    
    <h2>Интеграция и инфраструктура</h2>
    <ul>
      <li><strong>Конфигурация и мониторинг состояния:</strong> Использование <a href="https://www.consul.io/" target="_blank">Consul</a> для HealthCheck и хранения ключ-значение (KV).</li>
      <li><strong>Балансировка нагрузки:</strong> <a href="https://traefik.io/" target="_blank">Traefik</a> обеспечивает балансировку между серверами, получая список серверов и их флаги из Consul Catalog.</li>
      <li><strong>Хранение медиа:</strong> Фотографии сохраняются в объектном хранилище с помощью <a href="https://min.io/" target="_blank">Minio</a>.</li>
      <li><strong>Мониторинг:</strong> Сбор метрик осуществляется с помощью <a href="https://prometheus.io/" target="_blank">Prometheus</a>, а визуализация производится через <a href="https://grafana.com/" target="_blank">Grafana</a> (dashboard доступен в главной папке проекта).</li>
    </ul>
    
    <h2>План развития</h2>
    <ul>
      <li class="todo"><strong>TODO:</strong> Интеграция WebRTC для реализации голосовых и видео вызовов.</li>
      <li class="todo"><strong>TODO:</strong> Переход на gRPC вместо WebSockets для повышения производительности и надежности.</li>
    </ul>
    
    <p>Дальнейшая доработка проекта запланирована на ближайшее будущее.</p>
  </div>
</body>
</html>
```
