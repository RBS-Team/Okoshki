---
name: Registration architecture direction
description: users module owns registration endpoints; auth stays as pure credentials store
type: feedback
---

Registration is orchestrated by the users module, NOT auth.

**Why:** auth should stay clean — only credentials (CreateAccount, DeleteUser, Login, JWT). Registration (create account + save profile) belongs to users.

**How to apply:**
- `auth/service`: CreateAccount(email, password, role) → uuid, Login, DeleteAccount, GetUsersInfo. No profile logic.
- `users/service`: defines `AccountCreator` interface (duck typing, no auth import), calls it to create credentials, then saves profile. Compensating rollback on profile failure.
- Registration HTTP routes live in `users/delivery/http`, not in `auth/delivery/http`.
- `AccountCreator` interface is defined IN `users/service` (not in auth). auth/service implements it implicitly.
