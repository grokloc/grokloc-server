package app

const OrgsTableName = "orgs"
const UsersTableName = "users"
const RepositoriesTableName = "repositories"
const AuditTableName = "audit"

// Schema is the full schema to recreate the app db
const Schema = `
create table if not exists users (
       api_secret text unique not null,
       api_secret_digest text unique not null,
       id text unique not null,
       display_name text not null,
       display_name_digest text not null,
       email text not null,
       email_digest text not null,
       org text not null,
       password text not null,
       schema_version integer not null default 0,
       status integer not null,
       ctime integer,
       mtime integer,
       primary key (id));
-- STMT
create unique index if not exists users_email_org on users (email, org);
create unique index if not exists users_email_digest_org on users (email_digest, org);
-- STMT
create trigger if not exists users_ctime_trigger after insert on users
begin
        update users set
        ctime = strftime('%s','now'),
        mtime = strftime('%s','now')
        where id = new.id;
end;
-- STMT
create trigger if not exists users_mtime_trigger after update on users
begin
        update users set mtime = strftime('%s','now')
        where id = new.id;
end;
-- STMT
create table if not exists orgs (
       id text unique not null,
       name text unique not null,
       owner text not null,
       schema_version integer not null default 0,
       status integer not null,
       ctime integer,
       mtime integer,
       primary key (id));
-- STMT
create trigger if not exists orgs_ctime_trigger after insert on orgs
begin
        update orgs set
        ctime = strftime('%s','now'),
        mtime = strftime('%s','now')
        where id = new.id;
end;
-- STMT
create trigger if not exists orgs_mtime_trigger after update on orgs
begin
        update orgs set mtime = strftime('%s','now')
        where id = new.id;
end;
-- STMT
create table if not exists repositories (
       id text unique not null,
       name text not null,
       org text not null,
       path text not null,
       upstream text not null,
       schema_version integer not null default 0,
       status integer not null,
       ctime integer,
       mtime integer,
       primary key (id));
-- STMT
create unique index if not exists repositories_name_org on repositories (name, org);
-- STMT
create trigger if not exists repositories_ctime_trigger after insert on repositories
begin
        update repositories set
        ctime = strftime('%s','now'),
        mtime = strftime('%s','now')
        where id = new.id;
end;
-- STMT
create trigger if not exists repositories_mtime_trigger after update on repositories
begin
        update repositories set mtime = strftime('%s','now')
        where id = new.id;
end;
-- STMT
create table if not exists audit (
      id text unique not null,
      code integer not null,
      source text not null,
      source_id text not null,
      schema_version integer not null default 0,
      ctime integer,
      mtime integer,
      primary key (id));
-- STMT
create trigger if not exists audit_ctime_trigger after insert on audit
      begin
      update audit set
      ctime = strftime('%s','now'),
      mtime = strftime('%s','now')
      where id = new.id;
end;
-- STMT
create trigger if not exists audit_mtime_trigger after update on audit
      begin
      update audit set mtime = strftime('%s','now')
      where id = new.id;
end;
`
