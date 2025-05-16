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

<div class="container">
<h3>Основные функции</h3>
<ul>
<li>JWT аутентификация</li>
<li>Регистрация пользователей</li>
<li>WebSocket чат в реальном времени</li>
<li>Управление профилями (MinIO для аватарок)</li>
<li>Система блокировок</li>
<li>Метрики Prometheus</li>
</ul>

<h3>Технологии</h3>
<table>
<tr><td>Язык</td><td>Go</td></tr>
<tr><td>Фреймворк</td><td>Gorilla Mux</td></tr>
<tr><td>База данных</td><td>PostgreSQL (GORM)</td></tr>
<tr><td>Файловое хранилище</td><td>MinIO</td></tr>
</table>

<h3>API Endpoints</h3>
<div class="endpoint">
<strong>Аутентификация:</strong><br>
• <code>POST /service/api/login</code><br>
• <code>POST /service/api/register</code><br><br>

<strong>Чат:</strong><br>
• <code>GET /service/api/chats</code><br>
• <code>POST /service/api/chat</code><br><br>

<strong>Пользователи:</strong><br>
• <code>PUT /service/api/profile</code><br>
• <code>POST /service/api/profile/photo</code>
</div>

<h3>Метрики сервера</h3>
<ul>
<li>Время обработки сообщений</li>
<li>Количество активных чатов</li>
<li>Счётчик ошибок</li>
<li>Время работы сервера</li>
</ul>

<h3>Безопасность</h3>
<ul>
<li>CORS с ограниченными origin</li>
<li>Проверка блокировок перед отправкой сообщений</li>
<li>Валидация JWT токенов</li>
</ul>
</div>
