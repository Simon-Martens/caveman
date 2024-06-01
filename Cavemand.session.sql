INSERT INTO __users (
    email,
    name,
    password,
    created_on,
    modified_on,
    expires_on
  )
VALUES (
    'email:TEXT',
    'name:TEXT',
    'password:BLOB',
    created_on:INTEGER,
    modified_on:INTEGER,
    expires_on:INTEGER
  );