# Безопасность

## Секреты

Запрещено хранить в Git:
- GitHub tokens;
- API keys;
- пароли;
- приватные ключи;
- JWT secrets;
- платёжные credentials.

## Проверка

```bash
git grep -n -I -E 'ghp_|github_pat_|sk-|AKIA|PRIVATE KEY|password|secret'
```

Для истории рекомендуется secret-scanning инструмент.

Секреты хранятся в `.env` или защищённых GitHub Secrets.

## AI

AI-агентам не передаются паспортные данные, платёжные реквизиты, приватные договоры и токены.

Production deploy — только из защищённой ветки через PR.
