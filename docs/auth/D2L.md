## Auth & Context

### Bearer Token

D2L stores the session token in `localStorage` under `D2L.Fetch.Tokens`. The key `"*:*:*"` is the wildcard scope used for most API calls.

```javascript
const token = JSON.parse(localStorage["D2L.Fetch.Tokens"])["*:*:*"]
  .access_token;
// Pass as: Authorization: `Bearer ${token}`
```

Most `/d2l/api/le/` and `/d2l/api/lp/` endpoints require this. Exception: legacy endpoints like `/d2l/lms/` use cookie auth (session is sent automatically by the browser).

### `data-global-context`

D2L injects a JSON blob into the `<html>` element on every page. It contains the current user and course context.

```javascript
const ctx = JSON.parse(
  document.documentElement.getAttribute("data-global-context"),
);
// ctx.orgUnitId  — the current course/org unit ID ({ou} in endpoints)
// ctx.userId     — the current user's ID ({uid} in endpoints)
```

## /d2l/api/lp/ — Learning Platform

The core platform layer. User accounts, org structure, authentication, roles, enrollments. Things that exist regardless of whether you're in a course.

- `/lp/users/` — user accounts
- `/lp/enrollments/` — what courses a user is in
- `/lp/roles/` — role definitions
- `/lp/orgstructure/` — org unit hierarchy
- `/lp/auth/` — authentication

Think of it as the admin/identity layer.

## /d2l/api/le/ — Learning Environment

The course content layer. Everything that lives inside a specific course: assignments, grades, quizzes, content, classlist. Always scoped to an `orgUnitId`.

- `/le/{ver}/{ouId}/dropbox/` — assignments
- `/le/{ver}/{ouId}/grades/` — grades
- `/le/{ver}/{ouId}/content/toc` — content tree
- `/le/{ver}/{ouId}/classlist/` — students
- `/le/{ver}/{ouId}/news/` — announcements

# D2L Key Endpoints

| What              | Endpoint                                                        | Auth   |
| ----------------- | --------------------------------------------------------------- | ------ |
| Assignments list  | `GET /d2l/api/le/1.67/{ouId}/dropbox/folders/`                  | Bearer |
| Single assignment | `GET /d2l/api/le/1.67/{ouId}/dropbox/folders/{id}`              | Bearer |
| Submissions       | `GET /d2l/api/le/1.67/{ouId}/dropbox/folders/{id}/submissions/` | Bearer |
| Activity feed     | `GET /d2l/api/hm/activity?ou={ouId}`                            | Bearer |
| Legacy dropbox    | `GET /d2l/lms/dropbox/user/folders_history.d2l?ou={ouId}`       | Cookie |
| Quota check       | `GET /d2l/api/le/unstable/consumption/quota`                    | Bearer |
| Grades            | `GET /d2l/api/le/1.67/{ouId}/grades/`                           | Bearer |
| Course content    | `GET /d2l/api/le/1.67/{ouId}/content/toc`                       | Bearer |
| My enrollments    | `GET /d2l/api/lp/1.30/enrollments/myenrollments/`               | Bearer |
| User roles        | `GET /d2l/api/lp/1.30/users/{uid}/roles/`                       | Bearer |
| Class list        | `GET /d2l/api/le/1.67/{ouId}/classlist/`                        | Bearer |
| Course news       | `GET /d2l/api/le/1.67/{ouId}/news/`                             | Bearer |
| All Courses       | `GET /d2l/api/lp/1.30/enrollments/myenrollments/?userId=${uid}` | Bearer |

---

## Grades (extended)

| Method | Endpoint                                              | Description        |
| ------ | ----------------------------------------------------- | ------------------ |
| GET    | `/d2l/api/le/1.67/{ou}/grades/final/values/{uid}`     | Final grade        |
| GET    | `/d2l/api/le/1.67/{ou}/grades/{gradeId}/values/{uid}` | Single grade value |
| GET    | `/d2l/api/le/1.67/{ou}/grades/schemas/`               | Grading schemes    |

## Assignments (extended)

| Method | Endpoint                                                    | Description      |
| ------ | ----------------------------------------------------------- | ---------------- |
| GET    | `/d2l/api/le/1.67/{ou}/dropbox/folders/{id}/feedback/{uid}` | Feedback / grade |
| POST   | `/d2l/api/le/1.67/{ou}/dropbox/folders/{id}/submissions/`   | Submit file      |

## Quizzes

| Method | Endpoint                                          | Description       |
| ------ | ------------------------------------------------- | ----------------- |
| GET    | `/d2l/api/le/1.67/{ou}/quizzes/`                  | All quizzes       |
| GET    | `/d2l/api/le/1.67/{ou}/quizzes/{id}`              | Single quiz       |
| GET    | `/d2l/api/le/1.67/{ou}/quizzes/{id}/attempts/`    | Your attempts     |
| GET    | `/d2l/api/le/1.67/{ou}/quizzes/{id}/submissions/` | Submitted answers |

## Content (extended)

| Method | Endpoint                                                 | Description            |
| ------ | -------------------------------------------------------- | ---------------------- |
| GET    | `/d2l/api/le/1.67/{ou}/content/modules/{id}`             | Single module          |
| GET    | `/d2l/api/le/1.67/{ou}/content/topics/{id}`              | Single topic           |
| GET    | `/d2l/api/le/1.67/{ou}/content/topics/{id}/file`         | Download topic file    |
| GET    | `/d2l/api/le/1.67/{ou}/content/completions/`             | Your completion status |
| POST   | `/d2l/api/le/1.67/{ou}/content/topics/{id}/completions/` | Mark complete          |

## News (extended)

| Method | Endpoint                                            | Description         |
| ------ | --------------------------------------------------- | ------------------- |
| GET    | `/d2l/api/le/1.67/{ou}/news/{id}`                   | Single announcement |
| GET    | `/d2l/api/le/1.67/{ou}/news/{id}/attachments/{fid}` | Attachment          |

## Discussions

| Method | Endpoint                                                             | Description |
| ------ | -------------------------------------------------------------------- | ----------- |
| GET    | `/d2l/api/le/1.67/{ou}/discussions/forums/`                          | Forums      |
| GET    | `/d2l/api/le/1.67/{ou}/discussions/forums/{fid}/topics/`             | Threads     |
| GET    | `/d2l/api/le/1.67/{ou}/discussions/forums/{fid}/topics/{tid}/posts/` | Posts       |

## Calendar

| Method | Endpoint                                     | Description                 |
| ------ | -------------------------------------------- | --------------------------- |
| GET    | `/d2l/api/le/1.67/{ou}/calendar/events/`     | Course events               |
| GET    | `/d2l/api/lp/1.30/calendar/events/upcoming/` | All upcoming across courses |

## Users (extended)

| Method | Endpoint                                                      | Description  |
| ------ | ------------------------------------------------------------- | ------------ |
| GET    | `/d2l/api/lp/1.30/users/whoami`                               | Your profile |
 a

## Course Info

| Method | Endpoint                                        | Description                  |
| ------ | ----------------------------------------------- | ---------------------------- |
| GET    | `/d2l/api/lp/1.30/orgstructure/{ou}`            | Org unit details             |
| GET    | `/d2l/api/lp/1.30/orgstructure/{ou}/ancestors/` | Parent orgs (dept, semester) |
| GET    | `/d2l/api/lp/1.30/orgstructure/{ou}/children/`  | Child orgs                   |

## Surveys

| Method | Endpoint                                          | Description    |
| ------ | ------------------------------------------------- | -------------- |
| GET    | `/d2l/api/le/1.67/{ou}/surveys/`                  | All surveys    |
| GET    | `/d2l/api/le/1.67/{ou}/surveys/{id}/submissions/` | Your responses |

## Notifications / Alerts

| Method | Endpoint                                                             | Description                |
| ------ | -------------------------------------------------------------------- | -------------------------- |
| GET    | `/d2l/api/lp/1.30/notifications/instant/`                            | Instant notifications      |
| GET    | `/d2l/NavigationArea/{ou}/ActivityFeed/GetAlertsDaylight?Category=0` | Activity feed (category 0) |
| GET    | `/d2l/NavigationArea/{ou}/ActivityFeed/GetAlertsDaylight?Category=1` | Activity feed (category 1) |
| GET    | `/d2l/NavigationArea/{ou}/ActivityFeed/GetAlertsDaylight?Category=2` | Activity feed (category 2) |

## Locker

| Method | Endpoint                                         | Description      |
| ------ | ------------------------------------------------ | ---------------- |
| GET    | `/d2l/api/le/1.67/{ou}/locker/user/{uid}/`       | Your locker root |
| GET    | `/d2l/api/le/1.67/{ou}/locker/user/{uid}/{path}` | Specific folder  |

---

## Batch Test Script

```javascript
const token = JSON.parse(localStorage["D2L.Fetch.Tokens"])["*:*:*"]
  .access_token;
const ctx = JSON.parse(
  document.documentElement.getAttribute("data-global-context"),
);
const ou = ctx.orgUnitId,
  uid = ctx.userId;

const endpoints = [
  `/d2l/api/le/1.67/${ou}/grades/`,
  `/d2l/api/le/1.67/${ou}/dropbox/folders/`,
  `/d2l/api/le/1.67/${ou}/quizzes/`,
  `/d2l/api/le/1.67/${ou}/content/toc`,
  `/d2l/api/le/1.67/${ou}/news/`,
  `/d2l/api/le/1.67/${ou}/discussions/forums/`,
  `/d2l/api/lp/1.30/calendar/events/upcoming/`,
  `/d2l/api/lp/1.30/users/whoami`,
];

const results = await Promise.all(
  endpoints.map(async (url) => {
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const text = await res.text();
    let body;
    try {
      body = JSON.parse(text);
    } catch {
      body = text;
    }
    return { url, status: res.status, body };
  }),
);

results.forEach((r) => {
  if (r.status !== 200) {
    const icon = r.status === 403 ? "🔒" : "❌";
    console.warn(`${icon} ${r.status} — ${r.url}`, r.body);
    return;
  }
  console.group(`✅ ${r.status} — ${r.url}`);
  console.log(r.body);
  console.groupEnd();
});
```
