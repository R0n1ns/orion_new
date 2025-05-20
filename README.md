<h1>Архитектура чат-системы</h1>

<div class="container">
<h2>Основные особенности системы</h2>
<ul>
<li>WebSocket для сообщений, остальные запросы через HTTP</li>
<li>3-уровневая архитектура: Web Client → API Gateway → Server</li>
<li>Метрики Prometheus + Grafana: запросы, ошибки, сервер и gateway</li>
<li>Хранение данных: MinIO для файлов, PostgreSQL для структурированных данных</li>
</ul>
</div>

<h2>API Gateway</h2>

<div class="container">
<h3>Основные функции</h3>
<ul>
<li>Маршрутизация запросов</li>
<li>JWT аутентификация/авторизация</li>
<li>Rate Limiting (10 RPS)</li>
<li>Сбор метрик Prometheus</li>
<li>Проксирование запросов</li>
<li>Обслуживание статических страниц</li>
</ul>

<h3>Архитектура компонентов</h3>
<div class="endpoint">
<strong>Маршруты:</strong><br>
• Статические: <code>/</code>, <code>/chat</code>, <code>/login</code>, <code>/register</code><br>
• API: <code>/service/*</code><br>
• WebSocket: <code>/ws</code><br>
• Метрики: <code>/metrics</code>
</div>

<h3>Middleware</h3>
<ul>
<li>Проверка JWT</li>
<li>Rate Limiting:
    <ul>
        <li>Публичные: 10 RPS/IP</li>
        <li>Авторизованные: лимит по ID пользователя</li>
    </ul>
</li>
<li>Сбор метрик и логирование</li>
</ul>

<h3>Маршруты</h3>
<table>
<tr><th>Тип</th><th>Эндпоинты</th></tr>
<tr><td>Публичные</td><td><code>/login</code>, <code>/register</code>, <code>/metrics</code>, <code>/service/api/login</code>, <code>/service/api/register</code></td></tr>
<tr><td>Защищенные</td><td>Все остальные (требуют JWT)</td></tr>
</table>

<h3>Метрики</h3>
<div class="metrics">
<code>gateway_request_total</code> - общее количество запросов<br>
<code>gateway_request_duration_seconds</code> - время выполнения<br>
<code>gateway_user_request_total</code> - запросы по пользователям<br>
<code>gateway_rate_limit_blocked_total</code> - заблокированные запросы
</div>
</div>

<h2>Сервер</h2>

<h1>Архитектура сервера чат-системы</h1><div class="container"> <h2>Основные особенности сервера</h2> <ul> <li><strong>3-уровневая архитектура</strong>: Web Client → Server (прямое взаимодействие)</li> <li><strong>Технологии</strong>: <ul> <li>Gorilla Mux для маршрутизации HTTP</li> <li>GORM + PostgreSQL для хранения структурированных данных</li> <li>MinIO для хранения изображений профилей</li> <li>WebSocket (Gorilla) для реального времени</li> </ul> </li> <li><strong>Метрики</strong>: Prometheus + Grafana (время обработки, активные чаты, ошибки, аптайм)</li> <li><strong>Безопасность</strong>: JWT-аутентификация, CORS, блокировка пользователей</li> </ul> </div><h2>Основные функции сервера</h2> <div class="container"> <ul> <li>Управление пользователями: регистрация, блокировка, профиль</li> <li>Создание чатов (каналов) и управление ими</li> <li>Обмен сообщениями в реальном времени через WebSocket</li> <li>Хранение и отдача медиа через MinIO</li> <li>Сбор метрик производительности</li> <li>Обновление статусов онлайн/оффлайн</li> </ul> </div><h2>Архитектура компонентов</h2> <div class="container"> <div class="endpoint"> <strong>Маршруты (HTTP):</strong><br> • <code>/service/api/login</code> – аутентификация<br> • <code>/service/api/register</code> – регистрация<br> • <code>/service/api/chats</code> – управление чатами<br> • <code>/service/api/messages</code> – работа с сообщениями<br> • <code>/service/api/profile</code> – профиль пользователя<br> • <code>/service/api/block</code> – блокировка пользователей<br>
<strong>WebSocket:</strong>

• <code>/service/ws</code> – установка соединения для чатов

<strong>Метрики:</strong>

• <code>/service/metrics</code> – эндпоинт Prometheus

</div><h3>Схема взаимодействия</h3> <pre> Client → HTTP (REST) ├── User Management ├── Chat Operations └── Metrics
Client → WebSocket
├── Real-time Messages
└── Online Status Updates
</pre>

</div><h2>Middleware</h2> <div class="container"> <ul> <li><strong>CORS</strong>: Ограничение доменов, методов и заголовков</li> <li><strong>JWT Validation</strong>: Проверка токена для защищенных эндпоинтов</li> <li><strong>Metrics Collection</strong>: Автоматический сбор данных для Prometheus</li> </ul> </div><h2>Классификация маршрутов</h2> <div class="container"> <table> <tr><th>Тип</th><th>Эндпоинты</th></tr> <tr><td>Публичные</td><td><code>/api/login</code>, <code>/api/register</code>, <code>/metrics</code></td></tr> <tr><td>Защищенные</td><td>Все остальные (требуют JWT в заголовке)</td></tr> </table> </div><h2>Метрики Prometheus</h2> <div class="container metrics"> <code>app_request_total</code> – общее количество HTTP-запросов<br> <code>message_processing_time_seconds</code> – время обработки сообщений<br> <code>ws_manager_active_chats_total</code> – активные WebSocket-соединения<br> <code>app_error_total</code> – счетчик ошибок приложения<br> <code>app_uptime_seconds</code> – время работы сервера<br> <code>app_info</code> – информация о версии приложения<br> </div><h2>Особенности реализации</h2> <div class="container"> <ul> <li><strong>Блокировка пользователей</strong>: <ul> <li>Взаимная проверка блокировок перед отправкой сообщений</li> <li>Автоматическая разблокировка через фоновый worker</li> </ul> </li> <li><strong>Статусы онлайн</strong>: <ul> <li>Обновление через WebSocket-пинги каждые 30 сек</li> <li>Метрика активных чатов</li> </ul> </li> <li><strong>Оптимизация</strong>: <ul> <li>Кеширование Data URL для изображений профилей</li> <li>Batch-обработка сообщений в чатах</li> </ul> </li> </ul> </div>
