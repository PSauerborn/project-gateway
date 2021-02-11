--
-- PostgreSQL database dump
--

-- Dumped from database version 12.4 (Debian 12.4-1.pgdg100+1)
-- Dumped by pg_dump version 12.4 (Debian 12.4-1.pgdg100+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: app_configs; Type: TABLE; Schema: public; Owner: utils
--

CREATE TABLE public.app_configs (
    app_id uuid NOT NULL,
    app_name text NOT NULL,
    config json DEFAULT '{}'::json NOT NULL
);


ALTER TABLE public.app_configs OWNER TO utils;

--
-- Name: app_configs app_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: utils
--

ALTER TABLE ONLY public.app_configs
    ADD CONSTRAINT app_configs_pkey PRIMARY KEY (app_id);


--
-- PostgreSQL database dump complete
--