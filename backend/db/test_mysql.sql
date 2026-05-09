-- name: TestReturning :one
INSERT INTO PERSON (NAME) VALUES (?) RETURNING *;
