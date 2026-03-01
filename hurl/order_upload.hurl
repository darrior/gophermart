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

POST http://127.0.0.1:8080/api/user/orders
Content-Type: text/plain
[Cookies]
auth: $auth
```
61234
```

# GET http://127.0.0.1:8080/api/user/orders
# HTTP 200
