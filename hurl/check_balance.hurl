POST http://127.0.0.1:8080/api/user/register
{
  "login": "user1",
  "password": "123"
}

POST http://127.0.0.1:8080/api/user/login
{
  "login": "user1",
  "password": "123"
}
HTTP 200
[Captures]
auth: cookie "auth"

GET http://127.0.0.1:8080/api/user/balance
[Cookies]
auth: $auth
HTTP 200

