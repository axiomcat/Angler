# Migrations

## Seasons

```
alter table angle_tries add column season integer;
```

```
UPDATE angle_tries
SET season = 1
WHERE angle_issue < 1137;
```

```
UPDATE angle_tries
SET season = 2
WHERE angle_issue >= 1137;
UPDATE angle_tries
SET season = 2
WHERE angle_issue >= 1137;
```
