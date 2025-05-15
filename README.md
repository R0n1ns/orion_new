Чат и чат , чо такого

Основные моменты:
1) Отправка и получени сообщений по ws,все остальное по обычному http
2) Структура : web client --> api gateway --> server
3) Метрики , показывают количество запросов и ошибок с сервера и gateway: prometheus и grafana
4) Бд : минио для аватарок и postgre для всего остального

api gateway :
<h2>Основные функции</h2>
  <ul>
      <li>Маршрутизация запросов к соответствующим сервисам</li>
      <li>Аутентификация и авторизация пользователей через JWT</li>
      <li>Ограничение запросов (rate limiting) для защиты от перегрузки</li>
      <li>Сбор метрик работы системы в формате Prometheus</li>
      <li>Проксирование запросов к backend-сервисам</li>
      <li>Обслуживание статических страниц (чат, регистрация, вход)</li>
  </ul>
  
  <h2>Архитектура компонентов</h2>
  
  <h3>1. Маршрутизация</h3>
  <div class="endpoint">
      <strong>Статические маршруты:</strong><br>
      <code>/chat</code>, <code>/login</code>, <code>/register</code>, <code>/</code> - HTML-страницы<br>
      <strong>API Gateway:</strong><br>
      <code>/service/*</code> - перенаправление на backend-сервисы<br>
      <strong>WebSocket:</strong><br>
      <code>/ws</code> - обработка WebSocket соединений<br>
      <strong>Метрики:</strong><br>
      <code>/metrics</code> - endpoint для Prometheus
  </div>
  
  <h3>2. Middleware</h3>
  <p>Комбинированный middleware выполняет:</p>
  <ul>
      <li><strong>Аутентификацию:</strong> проверка JWT-токена для защищенных маршрутов</li>
      <li><strong>Rate Limiting:</strong>
          <ul>
              <li>Публичные запросы: ограничение по IP (10 RPS, burst до 10 запросов)</li>
              <li>Авторизованные запросы: ограничение по ID пользователя</li>
          </ul>
      </li>
      <li><strong>Метрики и логирование:</strong> сбор данных о всех запросах</li>
  </ul>
  
  <h3>3. Публичные и защищенные маршруты</h3>
  <p><strong>Публичные (не требуют аутентификации):</strong></p>
  <ul>
      <li><code>/service/api/register</code></li>
      <li><code>/service/api/login</code></li>
      <li><code>/login</code></li>
      <li><code>/register</code></li>
      <li><code>/metrics</code></li>
      <li><code>/chat</code></li>
      <li><code>/</code></li>
  </ul>
  
  <p><strong>Защищенные (требуют JWT-токен):</strong> все остальные маршруты</p>
  
  <h3>4. Собираемые метрики</h3>
  <div class="metrics">
      <p><code>gateway_request_total</code> - общее количество запросов (по методам и путям)</p>
      <p><code>gateway_request_duration_seconds</code> - время выполнения запросов</p>
      <p><code>gateway_user_request_total</code> - запросы по пользователям</p>
      <p><code>gateway_user_path_total</code> - запросы по путям для каждого пользователя</p>
      <p><code>gateway_rate_limit_blocked_total</code> - заблокированные запросы</p>
  </div>
  
  <h2>Особенности реализации</h2>
  <ul>
      <li>Поддержка WebSocket соединений</li>
      <li>Гибкая система rate limiting с разными профилями для публичных и авторизованных запросов</li>
      <li>Подробное логирование всех входящих запросов</li>
      <li>Интеграция с Prometheus для мониторинга</li>
      <li>Модульная архитектура для легкого расширения</li>
  </ul>

Сервер:

<h2>📋 Основные функции</h2>

<ul>
  <li><strong>Аутентификация</strong> (JWT)</li>
  <li><strong>Регистрация</strong> новых пользователей</li>
  <li><strong>WebSocket</strong> для реального времени</li>
  <li><strong>Чат-функционал</strong>:
    <ul>
      <li>Создание чатов</li>
      <li>Обмен сообщениями</li>
      <li>Отметки прочтения</li>
    </ul>
  </li>
  <li><strong>Профили пользователей</strong>:
    <ul>
      <li>Загрузка аватарок (MinIO)</li>
      <li>Редактирование данных</li>
    </ul>
  </li>
  <li><strong>Блокировка пользователей</strong></li>
  <li><strong>Метрики Prometheus</strong> для мониторинга</li>
</ul>

<h2>🛠 Технологии</h2>

<table>
  <tr>
    <td><strong>Язык</strong></td>
    <td>Go</td>
  </tr>
  <tr>
    <td><strong>Фреймворк</strong></td>
    <td>Gorilla Mux</td>
  </tr>
  <tr>
    <td><strong>База данных</strong></td>
    <td>PostgreSQL (через GORM)</td>
  </tr>
  <tr>
    <td><strong>Хранение файлов</strong></td>
    <td>MinIO</td>
  </tr>
  <tr>
    <td><strong>Реальное время</strong></td>
    <td>WebSocket (Gorilla)</td>
  </tr>
  <tr>
    <td><strong>Метрики</strong></td>
    <td>Prometheus</td>
  </tr>
  <tr>
    <td><strong>Безопасность</strong></td>
    <td>CORS, JWT</td>
  </tr>
</table>

<h2>🚀 API Endpoints</h2>

<h3>🔐 Аутентификация</h3>
<ul>
  <li><code>POST /service/api/login</code> - вход</li>
  <li><code>POST /service/api/register</code> - регистрация</li>
</ul>

<h3>💬 Чат</h3>
<ul>
  <li><code>GET /service/api/chats</code> - список чатов</li>
  <li><code>POST /service/api/chat</code> - создать чат</li>
  <li><code>GET /service/api/messages</code> - получить сообщения</li>
  <li><code>POST/PUT /service/api/messages/read</code> - отметить как прочитанные</li>
</ul>

<h3>👤 Пользователи</h3>
<ul>
  <li><code>GET /service/api/users</code> - поиск пользователей</li>
  <li><code>PUT /service/api/profile</code> - обновить профиль</li>
  <li><code>POST /service/api/profile/photo</code> - загрузить аватар</li>
</ul>

<h3>🚫 Блокировки</h3>
<ul>
  <li><code>POST /service/api/block</code> - заблокировать</li>
  <li><code>POST /service/api/unblock</code> - разблокировать</li>
  <li><code>GET /service/api/block-status</code> - статус блокировки</li>
  <li><code>GET /service/api/mutual-block</code> - взаимная блокировка</li>
</ul>

<h3>📊 Метрики</h3>
<ul>
  <li><code>GET /service/metrics</code> - Prometheus metrics</li>
</ul>

<h3>🔌 WebSocket</h3>
<ul>
  <li><code>GET /service/ws</code> - подключение WS</li>
</ul>

<h2>📈 Метрики</h2>
<p>Сервер предоставляет следующие метрики Prometheus:</p>
<ul>
  <li>Время обработки сообщений</li>
  <li>Количество активных чатов</li>
  <li>Счётчик запросов</li>
  <li>Время выполнения запросов</li>
  <li>Счётчик ошибок</li>
  <li>Время работы сервера</li>
</ul>

<h2>🔒 Безопасность</h2>
<ul>
  <li>CORS с ограниченными origin</li>
  <li>JWT аутентификация</li>
  <li>Проверка блокировок перед отправкой сообщений</li>
</ul>
