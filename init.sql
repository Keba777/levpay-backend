-- init.sql
create schema if not exists auth;
create schema if not exists app;
create schema if not exists user;
create schema if not exists wallet;
create schema if not exists transaction;
create schema if not exists kyc;
create schema if not exists file;
create schema if not exists notification;
create schema if not exists billing;
create schema if not exists admin;

create extension if not exists "uuid-ossp";
create extension if not exists pg_trgm;
create extension if not exists pgcrypto;