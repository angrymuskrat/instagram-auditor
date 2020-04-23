sudo -u postgres psql -c "CREATE DATABASE insta_audit;"
sudo -u postgres psql -d insta_audit -c "CREATE TABLE users( \
  id VARCHAR(15) PRIMARY KEY,\
  posts INT,\
  followers INT,\
  following INT,\
  type VARCHAR(20),\
  name TEXT,\
  alternate_name,\
  description, TEXT,\
)"
sudo -u postgres psql -d insta_audit -c "CREATE TABLE posts( \
\
)"