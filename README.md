<body>
  <h1>Проект: Реализация Чата</h1>
  <p><strong>Описание:</strong> Данный пет-проект представляет собой веб-чат с бэкендом на Go, использующим WebSocket для обмена данными. Фронтенд выполнен на HTML, CSS и JavaScript.</p>
  
  <h2>Функциональность</h2>
  <ul>
      <li>Взаимодействие между клиентом и сервером через WebSocket и JSON.</li>
      <li>Мониторинг работоспособности (HealthCheck) и хранение ключевых значений осуществляется через <strong>Consul</strong>.</li>
      <li>Балансировка нагрузки между серверами реализована через <strong>Traefik</strong>, который получает данные о серверах и флагах из <strong>Consul Catalog</strong>.</li>
      <li>Хранение изображений осуществляется в объектном хранилище с помощью <strong>Minio</strong>.</li>
      <li>Сбор и мониторинг метрик реализован с помощью <strong>Prometheus</strong>, с последующей визуализацией через <strong>Grafana</strong> (дэшборд находится в корневой папке проекта).</li>
  </ul>
  
  <h2>Планы на будущее (TODO)</h2>
  <ul>
      <li class="todo">Добавление поддержки WebRTC для голосовых и видеозвонков.</li>
      <li class="todo">Переписывание WebSocket на gRPC для более эффективной работы.</li>
  </ul>
  
  <p><strong>Прогнозируемая дата завершения:</strong> Завтра 😊</p>
